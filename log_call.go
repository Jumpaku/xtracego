package xtracego

import (
	"go/ast"

	"golang.org/x/tools/go/ast/astutil"
)

func (s *Xtrace) newCallLogStmt() ast.Stmt {
	// PrintlnCall("[CALL] <signature>")
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: ast.NewIdent(s.IdentifierPrintlnCall()),
			Args: []ast.Expr{
				&ast.Ident{Name: s.IdentShowTimestamp()},
				&ast.Ident{Name: s.IdentShowGoroutine()},
			},
		},
	}
}

func (s *Xtrace) newReturnLogStmt() ast.Stmt {
	// PrintlnReturn("[RETURN] <signature>")
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: ast.NewIdent(s.IdentifierPrintlnReturn()),
			Args: []ast.Expr{
				&ast.Ident{Name: s.IdentShowTimestamp()},
				&ast.Ident{Name: s.IdentShowGoroutine()},
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
		[]ast.Stmt{
			s.newCallLogStmt(),
			s.newReturnLogStmt(),
		},
		body.List...,
	)
	c.Replace(body)
	s.libraryRequired = true
}
