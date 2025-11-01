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

func ProcessCode(prefix, filename string, dst io.Writer, src io.Reader) (err error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}
	src, srcBytes := buf, buf.Bytes()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	astutil.AddNamedImport(fset, f, prefix+"_log", "log")
	astutil.AddNamedImport(fset, f, prefix+"_fmt", "fmt")

	s := LogInsert{fset: fset, src: srcBytes, lineWidth: 80, prefix: prefix}

	stmtParent := map[ast.Stmt]ast.Node{}
	ast.PreorderStack(f, nil, func(n ast.Node, s []ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil && len(node.Body.List) > 0 {
				stmtParent[node.Body.List[0]] = node
			}
		case *ast.ForStmt:
			if node.Body != nil && len(node.Body.List) > 0 {
				stmtParent[node.Body.List[0]] = node
			}
		case *ast.RangeStmt:
			if node.Body != nil && len(node.Body.List) > 0 {
				stmtParent[node.Body.List[0]] = node
			}
		case *ast.CaseClause:
			if node.Body != nil && len(node.Body) > 0 {
				for _, stmt := range node.Body {
					stmtParent[stmt] = node
				}
			}
		case *ast.CommClause:
			if node.Body != nil && len(node.Body) > 0 {
				for _, stmt := range node.Body {
					stmtParent[stmt] = node
				}
			}
		}
		return true
	})

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
		case ast.Stmt:
			{
				if parent, ok := stmtParent[node]; ok {
					switch parent := parent.(type) {
					case *ast.FuncDecl:
						s.logCall(c, parent)
					case *ast.ForStmt:
						s.logForInit(c, parent)
					case *ast.RangeStmt:
						s.logRangeKeyVal(c, parent)
					case *ast.CaseClause:
						s.logCaseClause(c, parent)
					case *ast.CommClause:
						s.logCommClause(c, parent)
					}
				}
			}

			switch node := node.(type) {
			case *ast.EmptyStmt:
			case *ast.BlockStmt:
			case *ast.DeclStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}

				if decl, ok := node.Decl.(*ast.GenDecl); ok && decl.Tok == token.VAR {
					s.logLocalVariable(c, node)
				}
			case *ast.AssignStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}

				if node.Tok == token.ASSIGN {
					s.logLocalAssignment(c, node)
				}
				if _, ok := c.Parent().(*ast.BlockStmt); ok && node.Tok == token.DEFINE {
					s.logLocalAssignment(c, node)
				}
			case *ast.ExprStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.IfStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.SwitchStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.TypeSwitchStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.CaseClause:
			case *ast.SelectStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.CommClause:
			case *ast.ForStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.RangeStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.ReturnStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.DeferStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.GoStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.BranchStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.LabeledStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.SendStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			case *ast.IncDecStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					s.logLocalStatement(c, node)
				}
			}
		}

		return true
	})

	if err := printer.Fprint(dst, fset, f); err != nil {
		return fmt.Errorf("failed to print: %w", err)
	}

	return nil
}
