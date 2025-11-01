package goxe

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"

	"golang.org/x/tools/go/ast/astutil"
)

func (s LogInsert) newCallLogStmt(name string) ast.Stmt {
	// log.Println(fmt.Sprintf(`[CALL] (reciever T) Method`))
	content := fmt.Sprintf("[CALL] %s", name)
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

func (s LogInsert) logCall(c *astutil.Cursor, funcDecl *ast.FuncDecl) {
	pattern := regexp.MustCompile(`\s+`)
	signature := s.fragment(funcDecl.Pos(), funcDecl.Body.Pos())
	signature = pattern.ReplaceAllString(signature, " ")
	c.InsertBefore(s.newCallLogStmt(fmt.Sprintf("%s", signature)))
}
