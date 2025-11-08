package xtracego

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/samber/lo/mutable"
	"golang.org/x/tools/go/ast/astutil"
)

func (s *Xtrace) newVariableLogStmt(name string, shadowed bool) ast.Stmt {
	// PrintlnVariable("VarName", VarName))
	// PrintlnVariable("VarName", "<shadowed>"))
	if shadowed {
		return &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: ast.NewIdent(s.IdentifierPrintlnVariable()),
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf("%q", name),
					},
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf("%q", "<shadowed>"),
					},
					&ast.Ident{Name: s.IdentShowTimestamp()},
					&ast.Ident{Name: s.IdentShowGoroutine()},
				},
			},
		}
	} else {
		return &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: ast.NewIdent(s.IdentifierPrintlnVariable()),
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf("%q", name),
					},
					&ast.Ident{Name: name},
					&ast.Ident{Name: s.IdentShowTimestamp()},
					&ast.Ident{Name: s.IdentShowGoroutine()},
				},
			},
		}
	}
}

func (s *Xtrace) newVariableLogDecl(name string, shadowed bool) *ast.GenDecl {
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
									s.newVariableLogStmt(name, shadowed),
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

func (s *Xtrace) logFileVariable(c *astutil.Cursor, node *ast.GenDecl) {
	if !s.TraceVar {
		return
	}

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
			decls = append(decls, s.newVariableLogDecl(name.Name, false))
		}
	}
	mutable.Reverse(decls)
	for _, decl := range decls {
		c.InsertAfter(decl)
		s.libraryRequired = true
	}

}

func (s *Xtrace) logLocalVariable(c *astutil.Cursor, node *ast.DeclStmt) {
	if !s.TraceVar {
		return
	}

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
			stmts = append(stmts, s.newVariableLogStmt(name.Name, false))
		}
	}
	mutable.Reverse(stmts)
	for _, decl := range stmts {
		c.InsertAfter(decl)
		s.libraryRequired = true
	}
}

func (s *Xtrace) logLocalAssignment(c *astutil.Cursor, node *ast.AssignStmt) {
	if !s.TraceVar {
		return
	}

	stmts := []ast.Stmt{}
	for _, lexpr := range node.Lhs {
		ident, ok := lexpr.(*ast.Ident)
		if !ok || ident.Name == "_" {
			continue
		}
		stmts = append(stmts, s.newVariableLogStmt(ident.Name, false))
	}
	mutable.Reverse(stmts)
	for _, decl := range stmts {
		c.InsertAfter(decl)
		s.libraryRequired = true
	}

}

func (s *Xtrace) logForVariables(c *astutil.Cursor, info *ForInfo) {
	if !s.TraceVar {
		return
	}

	body := info.Body
	vars := []ast.Stmt{}
	for _, ident := range info.Variables() {
		vars = append(vars, s.newVariableLogStmt(ident.Name, false))
	}
	body.List = append(vars, body.List...)
	c.Replace(body)
	s.libraryRequired = len(vars) > 0

}

func (s *Xtrace) logIfVariables(c *astutil.Cursor, info *IfElseInfo) {
	if !s.TraceVar {
		return
	}

	vars := info.Variables()
	shadow := map[string]bool{}
	stmts := []ast.Stmt{}
	for i := len(vars) - 1; i >= 0; i-- {
		stmts = append(stmts, s.newVariableLogStmt(vars[i].Name, shadow[vars[i].Name]))
		shadow[vars[i].Name] = true
	}
	mutable.Reverse(stmts)

	if info.Body != nil {
		info.Body.List = append(stmts, info.Body.List...)
		c.Replace(info.Body)
		s.libraryRequired = len(vars) > 0
	}
	if info.ElseBody != nil {
		info.ElseBody.List = append(stmts, info.ElseBody.List...)
		c.Replace(info.ElseBody)
		s.libraryRequired = len(vars) > 0
	}
}

func (s *Xtrace) logCallVariables(c *astutil.Cursor, info *FuncInfo) {
	if !s.TraceCall {
		return
	}
	params := []ast.Stmt{}
	for _, param := range info.FuncDecl.Type.Params.List {
		for _, name := range param.Names {
			if name.Name == "_" {
				continue
			}
			params = append(params, s.newVariableLogStmt(name.Name, false))
		}
	}
	info.Body.List = append(params, info.Body.List...)
	c.Replace(info.Body)
	s.libraryRequired = true
}
