package xtracego

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/tools/go/ast/astutil"
)

func (s *Xtrace) newStatementLogStmt(pos token.Position, fragment string) ast.Stmt {
	// log.Println(fmt.Sprintf(`if a == 1 { /* path/to/source.go:123:45 */`))
	content := fmt.Sprintf("%s ", fragment)
	content = strings.ReplaceAll(content, "\t", "    ")
	if len(content) < s.LineWidth {
		content += strings.Repeat(".", s.LineWidth-len(content))
	}
	content += fmt.Sprintf(" [ %s:%d:%d ]", pos.Filename, pos.Line, pos.Column)
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: ast.NewIdent(s.Prefix + "_log.Println"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent(s.Prefix + "_fmt.Sprintf"),
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", strings.ReplaceAll(content, "%", "%%")),
						},
					},
				},
			},
		},
	}
}

func (s *Xtrace) newStatementLogDecl(pos token.Position, fragment string) *ast.GenDecl {
	//var _ = func() int {
	//	log.Println(fmt.Sprintf(`if a == 1 { /* path/to/source.go:123:45 */`))
	//	return 0
	//}()
	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent("_")},
				Values: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.FuncLit{
							Type: &ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("int")}}}},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									s.newStatementLogStmt(pos, fragment),
									&ast.ReturnStmt{Results: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (s *Xtrace) logFileStatement(c *astutil.Cursor, node *ast.GenDecl) {
	if !s.TraceStmt {
		return
	}

	for _, spec := range node.Specs {
		spec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		pos := s.fset.Position(spec.Pos())
		frag := s.fragmentLine(spec.Pos())
		c.InsertBefore(s.newStatementLogDecl(pos, frag))
		s.requireImport = true
	}
}

func (s *Xtrace) tryLogLocalStatement(c *astutil.Cursor, node ast.Stmt) {
	if !s.TraceStmt {
		return
	}

	{
		insertable := false
		switch parent := c.Parent().(type) {
		case *ast.BlockStmt:
			if lo.Contains(parent.List, node) {
				insertable = true
			}
		case *ast.SwitchStmt:
		case *ast.TypeSwitchStmt:
		case *ast.SelectStmt:
		case *ast.ForStmt:
		case *ast.RangeStmt:
		case *ast.IfStmt:
		case *ast.CaseClause:
			if lo.Contains(parent.Body, node) {
				insertable = true
			}
		case *ast.CommClause:
			if lo.Contains(parent.Body, node) {
				insertable = true
			}
		case *ast.ReturnStmt:
		case *ast.DeferStmt:
		case *ast.GoStmt:
		case *ast.BranchStmt:
		case *ast.LabeledStmt:
		case *ast.SendStmt:
		case *ast.IncDecStmt:
		case *ast.ExprStmt:
		case *ast.AssignStmt:
		case *ast.EmptyStmt:
		}

		tracable := true
		switch node.(type) {
		case *ast.BlockStmt:
			tracable = false
		case *ast.SwitchStmt:
		case *ast.TypeSwitchStmt:
		case *ast.SelectStmt:
		case *ast.ForStmt:
		case *ast.RangeStmt:
		case *ast.IfStmt:
			if _, ok := c.Parent().(*ast.IfStmt); ok {
				tracable = false
			}
		case *ast.CaseClause:
			tracable = false
		case *ast.CommClause:
			tracable = false
		case *ast.ReturnStmt:
		case *ast.DeferStmt:
		case *ast.GoStmt:
		case *ast.BranchStmt:
		case *ast.LabeledStmt:
		case *ast.SendStmt:
		case *ast.IncDecStmt:
		case *ast.ExprStmt:
		case *ast.AssignStmt:
		case *ast.EmptyStmt:
			tracable = false
		}

		if !insertable || !tracable {
			return
		}
	}

	pos := s.fset.Position(node.Pos())
	frag := s.fragmentLine(node.Pos())
	c.InsertBefore(s.newStatementLogStmt(pos, frag))
	s.requireImport = true
}

func (s *Xtrace) logIfElseStatement(c *astutil.Cursor, info *IfElseInfo) {
	if !s.TraceStmt {
		return
	}

	stmts := []ast.Stmt{}
	for i, ifStmt := range info.Parents {
		if i == 0 {
			continue
		}
		frag := s.fragmentLine(ifStmt.If)
		stmts = append(stmts, s.newStatementLogStmt(s.fset.Position(ifStmt.If), frag))
	}
	if len(info.Parents) > 0 {
		frag := s.fragmentLine(info.IfStmt.If)
		stmts = append(stmts, s.newStatementLogStmt(s.fset.Position(info.IfStmt.If), frag))
		if info.ElseBody != nil {
			frag := s.fragmentLine(info.IfStmt.Body.Rbrace)
			stmts = append(stmts, s.newStatementLogStmt(s.fset.Position(info.IfStmt.Body.Rbrace), frag))
		}
	}
	if info.Body != nil {
		info.Body.List = append(stmts, info.Body.List...)
		c.Replace(info.Body)
		s.requireImport = len(stmts) > 0
	}
	if info.ElseBody != nil {
		info.ElseBody.List = append(stmts, info.ElseBody.List...)
		c.Replace(info.ElseBody)
		s.requireImport = len(stmts) > 0
	}
}
