package xtracego

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

func (s *Xtrace) newCallLogStmt(name string) ast.Stmt {
	// log.Println(fmt.Sprintf(`[CALL] (reciever T) Method`))
	content := fmt.Sprintf("[CALL] %s", name)
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: ast.NewIdent(s.Prefix + "_log.Println"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent(s.Prefix + "_fmt.Sprintf"),
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", content),
						},
					},
				},
			},
		},
	}
}

func (s *Xtrace) logCall(c *astutil.Cursor, info *FuncInfo) {
	if !s.TraceCall {
		return
	}
	body := info.Body
	body.List = append(
		[]ast.Stmt{s.newCallLogStmt(s.fragment(info.Signature()))},
		body.List...,
	)
	c.Replace(body)
	s.requireImport = true
}
