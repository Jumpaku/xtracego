package xtracego

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/samber/lo/mutable"
	"golang.org/x/tools/go/ast/astutil"
)

func (s *Xtrace) newVariableLogStmt(stack int, name string, shadowed bool) ast.Stmt {
	// PrintlnVariable("VarName", VarName))
	// PrintlnVariable("VarName", "<shadowed>"))
	if shadowed {
		return &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: ast.NewIdent(s.IdentifierPrintlnVariable()),
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.INT,
						Value: fmt.Sprintf(`%d`, stack),
					},
					&ast.BasicLit{
						Kind:  token.INT,
						Value: fmt.Sprintf(`%d`, s.LineWidth),
					},
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf("%q", name),
					},
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: `"<shadowed>"`,
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
						Kind:  token.INT,
						Value: fmt.Sprintf(`%d`, stack),
					},
					&ast.BasicLit{
						Kind:  token.INT,
						Value: fmt.Sprintf(`%d`, s.LineWidth),
					},
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
func (s *Xtrace) newReturnVariableLogStmt(number int, name string) ast.Stmt {
	// defer func () { PrintlnVariable("VarName", VarName)) }
	// defer func () { PrintlnVariable("<return_1>", return_1_abcdefg)) }
	varName := name
	if name == "" {
		name = fmt.Sprintf("<return_%d>", number)
		varName = fmt.Sprintf("return_%d_%s", number, s.UniqueString)
	}
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Type: &ast.FuncType{},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: ast.NewIdent(s.IdentifierPrintlnReturnVariable()),
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.INT,
										Value: fmt.Sprintf(`%d`, 4),
									},
									&ast.BasicLit{
										Kind:  token.INT,
										Value: fmt.Sprintf(`%d`, s.LineWidth),
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: fmt.Sprintf("%q", name),
									},
									&ast.Ident{Name: varName},
									&ast.Ident{Name: s.IdentShowTimestamp()},
									&ast.Ident{Name: s.IdentShowGoroutine()},
								},
							},
						},
					},
				},
			},
		},
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
									s.newVariableLogStmt(4, name, shadowed),
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
			stmts = append(stmts, s.newVariableLogStmt(3, name.Name, false))
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
		stmts = append(stmts, s.newVariableLogStmt(3, ident.Name, false))
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
		vars = append(vars, s.newVariableLogStmt(3, ident.Name, false))
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
		stmts = append(stmts, s.newVariableLogStmt(3, vars[i].Name, shadow[vars[i].Name]))
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
	fields := []*ast.Field{}
	if info.FuncDecl != nil {
		fields = info.FuncDecl.Type.Params.List
	}
	if info.FuncLit != nil {
		fields = info.FuncLit.Type.Params.List
	}

	params := []ast.Stmt{}
	for _, param := range fields {
		for _, name := range param.Names {
			if name.Name == "_" {
				continue
			}
			params = append(params, s.newVariableLogStmt(3, name.Name, false))
		}
	}

	info.Body.List = append(params, info.Body.List...)
	c.Replace(info.Body)
	s.libraryRequired = true
}

func (s *Xtrace) logReturnVariables(c *astutil.Cursor, info *FuncInfo) {
	fields := []*ast.Field{}
	if info.FuncDecl != nil && info.FuncDecl.Type.Results != nil {
		fields = info.FuncDecl.Type.Results.List
	}
	if info.FuncLit != nil && info.FuncLit.Type.Results != nil {
		fields = info.FuncLit.Type.Results.List
	}

	params := []ast.Stmt{}
	count := 0
	for _, param := range fields {
		if len(param.Names) == 0 {
			count++
			params = append(params, s.newReturnVariableLogStmt(count, ""))
		} else {
			for _, name := range param.Names {
				varName := name.Name
				if varName == "_" {
					varName = ""
				}
				count++
				params = append(params, s.newReturnVariableLogStmt(count, varName))
			}
		}
	}

	slices.Reverse(params)
	info.Body.List = append(params, info.Body.List...)
	c.Replace(info.Body)
	s.libraryRequired = true
}
