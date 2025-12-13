package xtracego

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func (s *Xtrace) newCallLogStmt(signature string) ast.Stmt {
	// PrintlnCall("[CALL] <signature>")
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: ast.NewIdent(s.IdentifierPrintlnCall()),
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.INT,
					Value: fmt.Sprintf(`%d`, s.LineWidth),
				},
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`%q`, signature),
				},
				&ast.Ident{Name: s.IdentShowTimestamp()},
				&ast.Ident{Name: s.IdentShowGoroutine()},
			},
		},
	}
}

func (s *Xtrace) newReturnLogStmt(signature string) ast.Stmt {
	// PrintlnReturn("[RETURN] <signature>")
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: ast.NewIdent(s.IdentifierPrintlnReturn()),
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.INT,
					Value: fmt.Sprintf(`%d`, s.LineWidth),
				},
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`%q`, signature),
				},
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
	signature := strings.Join(strings.Fields(s.fragment(info.Signature())), " ")
	body := info.Body
	body.List = append(
		[]ast.Stmt{
			s.newCallLogStmt(signature),
			s.newReturnLogStmt(signature),
		},
		body.List...,
	)
	c.Replace(body)
	s.libraryRequired = true
}
