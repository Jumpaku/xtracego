package xtracego

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"

	"golang.org/x/tools/go/ast/astutil"
)

func (s Xtrace) newCaseLogStmt(clause string) ast.Stmt {
	// log.Println(fmt.Sprintf(`[CASE] case conditions:`))
	content := fmt.Sprintf("[CASE] %s", clause)
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: ast.NewIdent(s.prefix + "_log.Println"),
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent(s.prefix + "_fmt.Sprintf"),
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

func (s Xtrace) logCase(c *astutil.Cursor, info *CaseInfo) {
	pattern := regexp.MustCompile(`\s+`)
	frag := s.fragment(info.CaseLabel())
	frag = pattern.ReplaceAllString(frag, " ")
	stmt := s.newCaseLogStmt(fmt.Sprintf("%s:", frag))
	if info.Case != nil {
		info.Case.Body = append([]ast.Stmt{stmt}, info.Case.Body...)
		c.Replace(info.Case)
	}
	if info.Comm != nil {
		info.Comm.Body = append([]ast.Stmt{stmt}, info.Comm.Body...)
		c.Replace(info.Comm)
	}
}
