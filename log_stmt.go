package goxe

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

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
			Fun: ast.NewIdent(s.prefix + "_log.Println"),
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
