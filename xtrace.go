package xtracego

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"math/rand"

	"golang.org/x/tools/go/ast/astutil"
)

type Config struct {
	TraceStmt bool
	TraceVar  bool
	TraceCall bool
	TraceCase bool
	Prefix    string
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func (cfg *Config) GenPrefix(seed int64) {
	r := rand.New(rand.NewSource(seed))
	v := []byte{}
	for i := 0; i < 8; i++ {
		v = append(v, alphabet[r.Intn(len(alphabet))])
	}
	cfg.Prefix = "xtracego_" + string(v)
}

func ProcessCode(cfg Config, filename string, dst io.Writer, src io.Reader) (err error) {
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

	astutil.AddNamedImport(fset, f, cfg.Prefix+"_log", "log")
	astutil.AddNamedImport(fset, f, cfg.Prefix+"_fmt", "fmt")

	s := Xtrace{
		fset:      fset,
		src:       srcBytes,
		lineWidth: 80,
		prefix:    cfg.Prefix,

		funcByBody:   CollectFuncInfo(f),
		forByBody:    CollectForInfo(f),
		caseByBody:   CollectCaseInfo(f),
		ifElseByBody: CollectIfElseInfo(f),
	}

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
				s.logFileVariable(c, node)
			}
		case ast.Stmt:
			{
				if info, ok := s.funcByBody[node]; ok {
					s.logCall(c, info)
				}
				if info, ok := s.forByBody[node]; ok {
					s.logForVariables(c, info)
				}
				if info, ok := s.caseByBody[node]; ok {
					s.logCase(c, info)
				}
				if info, ok := s.ifElseByBody[node]; ok {
					s.logIfElseVariables(c, info)
					s.logIfElseStatement(c, info)
				}
			}

			s.tryLogLocalStatement(c, node)

			switch node := node.(type) {
			case *ast.DeclStmt:
				if decl, ok := node.Decl.(*ast.GenDecl); ok && decl.Tok == token.VAR {
					s.logLocalVariable(c, node)
				}
			case *ast.AssignStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					if node.Tok == token.ASSIGN || node.Tok == token.DEFINE {
						s.logLocalAssignment(c, node)
					}
				}
			case *ast.EmptyStmt:
			case *ast.BlockStmt:
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
		}

		return true
	})

	if err := printer.Fprint(dst, fset, f); err != nil {
		return fmt.Errorf("failed to print: %w", err)
	}

	return nil
}
