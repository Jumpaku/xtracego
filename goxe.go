package goxe

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"

	"golang.org/x/tools/go/ast/astutil"
)

func ProcessCode(filename string, dst io.Writer, src io.Reader) (err error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}
	src, srcBytes := buf, buf.Bytes()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0 /*parser.ParseComments*/)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	astutil.AddImport(fset, f, "log")
	astutil.AddImport(fset, f, "fmt")

	s := LogInsert{fset: fset, src: srcBytes, lineWidth: 80}

	astutil.Apply(f, nil, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		case *ast.GenDecl:
			switch node.Tok {
			case token.VAR:
				if _, isFile := c.Parent().(*ast.File); isFile {
					s.logFileStatement(c, node)
					s.logFileVariable(c, node)
				}
			case token.CONST:
				s.logFileStatement(c, node)
				s.logFileConstant(c, node)
			}
		case *ast.BlockStmt, *ast.EmptyStmt:
			// do nothing
		case *ast.DeclStmt:
			if decl, ok := node.Decl.(*ast.GenDecl); ok && decl.Tok == token.VAR {
				s.logLocalStatement(c, node)
				s.logLocalVariable(c, node)
			}
		case *ast.AssignStmt:
			if node.Tok == token.ASSIGN {
				s.logLocalStatement(c, node)
				s.logLocalAssignment(c, node)
			}
			if _, ok := c.Parent().(*ast.BlockStmt); ok && node.Tok == token.DEFINE {
				s.logLocalStatement(c, node)
				s.logLocalAssignment(c, node)
			}
		case ast.Stmt:
			s.logLocalStatement(c, node)
		case *ast.ExprStmt:
		case *ast.IfStmt:
		case *ast.SwitchStmt:
		case *ast.TypeSwitchStmt:
		case *ast.CaseClause:
		case *ast.SelectStmt:
		case *ast.CommClause:
		case *ast.ForStmt:
		case *ast.RangeStmt:
		case *ast.ReturnStmt:
		case *ast.DeferStmt:
		case *ast.GoStmt:
		case *ast.BranchStmt:
		case *ast.LabeledStmt:
		case *ast.SendStmt:
		case *ast.IncDecStmt:
		}
		return true
	})

	if err := printer.Fprint(dst, fset, f); err != nil {
		return fmt.Errorf("failed to print: %w", err)
	}

	return nil
}
