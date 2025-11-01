package goxe

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"

	"golang.org/x/tools/go/ast/astutil"
)

func (s LogInsert) newCaseLogStmt(clause string) ast.Stmt {
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

func (s LogInsert) logCaseClause(c *astutil.Cursor, clause *ast.CaseClause) {
	pattern := regexp.MustCompile(`\s+`)
	frag := s.fragment(clause.Pos(), clause.Colon)
	frag = pattern.ReplaceAllString(frag, " ")
	c.InsertBefore(s.newCaseLogStmt(fmt.Sprintf("%s:", frag)))
}

func (s LogInsert) logCommClause(c *astutil.Cursor, clause *ast.CommClause) {
	pattern := regexp.MustCompile(`\s+`)
	frag := s.fragment(clause.Pos(), clause.Colon)
	frag = pattern.ReplaceAllString(frag, " ")
	c.InsertBefore(s.newCaseLogStmt(fmt.Sprintf("%s:", frag)))
}
