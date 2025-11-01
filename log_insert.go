package goxe

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/samber/lo/mutable"
	"golang.org/x/tools/go/ast/astutil"
)

type LogInsert struct {
	fset      *token.FileSet
	src       []byte
	lineWidth int
}

func (s LogInsert) newVariableLogStmt(pos token.Position, name string) ast.Stmt {
	// log.Println(fmt.Sprintf(`[variable] VarName: %+v path/to/source.go:123:45`, VarName))
	content := fmt.Sprintf("[VAR] %s=%%#v", name)
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: ast.NewIdent("log.Println"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("fmt.Sprintf"),
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", content),
						},
						&ast.Ident{Name: name},
					},
				},
			},
		},
	}
}

func (s LogInsert) newVariableLogDecl(pos token.Position, name string) *ast.GenDecl {
	//var _ = func() int {
	//	log.Println(fmt.Sprintf(`VarName: %+v path/to/source.go:123:45`, VarName))
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
									s.newVariableLogStmt(pos, name),
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

func (s LogInsert) newStatementLogStmt(pos token.Position, fragment string) ast.Stmt {
	// log.Println(fmt.Sprintf(`if a == 1 { /* path/to/source.go:123:45 */`))
	content := fmt.Sprintf("%s ", fragment)
	content = strings.ReplaceAll(content, "\t", "    ")
	if len(content) < s.lineWidth {
		content += strings.Repeat(".", s.lineWidth-len(content))
	}
	content += fmt.Sprintf(" [ %s:%d:%d ]", pos.Filename, pos.Line, pos.Column)
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: ast.NewIdent("log.Println"),
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("%q", content),
				},
			},
		},
	}
}

func (s LogInsert) newStatementLogDecl(pos token.Position, fragment string) *ast.GenDecl {
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

func (s LogInsert) logFileVariable(c *astutil.Cursor, node *ast.GenDecl) {
	decls := []*ast.GenDecl{}
	for _, spec := range node.Specs {
		spec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range spec.Names {
			if name.Name == "_" {
				continue
			}
			pos := s.fset.Position(name.Pos())
			decls = append(decls, s.newVariableLogDecl(pos, name.Name))
		}
	}
	mutable.Reverse(decls)
	for _, decl := range decls {
		c.InsertAfter(decl)
	}
}

func (s LogInsert) logFileConstant(c *astutil.Cursor, node *ast.GenDecl) {
	decls := []*ast.GenDecl{}
	for _, spec := range node.Specs {
		spec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range spec.Names {
			if name.Name == "_" {
				continue
			}
			pos := s.fset.Position(name.Pos())
			decls = append(decls, s.newVariableLogDecl(pos, name.Name))
		}
	}
	mutable.Reverse(decls)
	for _, decl := range decls {
		c.InsertAfter(decl)
	}
}

func (s LogInsert) logLocalVariable(c *astutil.Cursor, node *ast.DeclStmt) {
	decl, ok := node.Decl.(*ast.GenDecl)
	if !ok {
		return
	}
	stmts := []ast.Stmt{}
	for _, spec := range decl.Specs {
		spec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range spec.Names {
			if name.Name == "_" {
				continue
			}
			pos := s.fset.Position(name.Pos())
			stmts = append(stmts, s.newVariableLogStmt(pos, name.Name))
		}
	}
	mutable.Reverse(stmts)
	for _, decl := range stmts {
		c.InsertAfter(decl)
	}
}

func (s LogInsert) logLocalAssignment(c *astutil.Cursor, node *ast.AssignStmt) {
	stmts := []ast.Stmt{}
	for _, lexpr := range node.Lhs {
		ident, ok := lexpr.(*ast.Ident)
		if !ok || ident.Name == "_" {
			continue
		}
		pos := s.fset.Position(ident.Pos())
		stmts = append(stmts, s.newVariableLogStmt(pos, ident.Name))
	}
	mutable.Reverse(stmts)
	for _, decl := range stmts {
		c.InsertAfter(decl)
	}
}

func (s LogInsert) fragment(pos, end token.Pos) string {
	return string(s.src[pos-1 : end-1])
}
func (s LogInsert) fragmentLine(pos, end token.Pos) string {
	begin := pos - 1
	for ; begin > 0; begin-- {
		if s.src[begin-1] == '\n' || s.src[begin-1] == '\r' {
			break
		}
	}
	frag := s.fragment(begin+1, end)
	frag, _, _ = strings.Cut(frag, "\n")
	frag, _, _ = strings.Cut(frag, "\r")
	return frag
}

func (s LogInsert) logFileStatement(c *astutil.Cursor, node *ast.GenDecl) {
	for _, spec := range node.Specs {
		spec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		pos := s.fset.Position(spec.Pos())
		frag := s.fragmentLine(spec.Pos(), spec.End())
		c.InsertBefore(s.newStatementLogDecl(pos, frag))
	}
}

func (s LogInsert) logLocalStatement(c *astutil.Cursor, node ast.Stmt) {
	pos := s.fset.Position(node.Pos())
	frag := s.fragmentLine(node.Pos(), node.End())
	c.InsertBefore(s.newStatementLogStmt(pos, frag))
}
