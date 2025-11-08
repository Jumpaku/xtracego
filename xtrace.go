package xtracego

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"

	"github.com/samber/lo/mutable"
	"golang.org/x/tools/go/ast/astutil"
)

type Config struct {
	TraceStmt bool
	TraceVar  bool
	TraceCall bool

	ShowTimestamp bool
	ShowGoroutine bool

	UniqueString string
	LineWidth    int

	ResolveType ResolveType
	ModuleName  string
}

func (cfg *Config) LibraryPackageName() string {
	if cfg.ResolveType == ResolveType_CommandLineArguments {
		return "main"
	}
	return "xtracego_" + cfg.UniqueString
}

func (cfg *Config) LibraryImportPath() string {
	if cfg.ResolveType == ResolveType_CommandLineArguments {
		return ""
	}
	return cfg.ModuleName + "/" + cfg.LibraryPackageName()
}

func (cfg *Config) LibraryFileName() string {
	return "xtracego_" + cfg.UniqueString + ".go"
}

func (cfg *Config) ExecutableFileName() string {
	return "main_" + cfg.UniqueString
}

func (cfg *Config) IdentifierPrintlnStatement() string {
	funcName := "PrintlnStatement_" + cfg.UniqueString
	if cfg.ResolveType == ResolveType_CommandLineArguments {
		return funcName
	}
	return cfg.LibraryPackageName() + "." + funcName
}

func (cfg *Config) IdentifierPrintlnVariable() string {
	funcName := "PrintlnVariable_" + cfg.UniqueString
	if cfg.ResolveType == ResolveType_CommandLineArguments {
		return funcName
	}
	return cfg.LibraryPackageName() + "." + funcName
}

func (cfg *Config) IdentifierPrintlnCall() string {
	funcName := "PrintlnCall_" + cfg.UniqueString
	if cfg.ResolveType == ResolveType_CommandLineArguments {
		return funcName
	}
	return cfg.LibraryPackageName() + "." + funcName
}

func (cfg *Config) IdentifierPrintlnReturn() string {
	funcName := "PrintlnReturn_" + cfg.UniqueString
	if cfg.ResolveType == ResolveType_CommandLineArguments {
		return funcName
	}
	return cfg.LibraryPackageName() + "." + funcName
}

func ProcessCode(config Config, filename string, src []byte) (dst []byte, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	config.LineWidth = 80
	s := &Xtrace{
		Config: config,
		fset:   fset,
		src:    src,

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
					s.logCallVariables(c, info)
					s.logCall(c, info)
				}
				if info, ok := s.forByBody[node]; ok {
					s.logForVariables(c, info)
				}
				if info, ok := s.caseByBody[node]; ok {
					s.logCaseStatement(c, info)
				}
				if info, ok := s.ifElseByBody[node]; ok {
					s.logIfVariables(c, info)
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

	if s.libraryRequired && s.ResolveType != ResolveType_CommandLineArguments {
		astutil.AddImport(fset, f, config.LibraryImportPath())
	}

	buf := bytes.NewBuffer(nil)
	if err := printer.Fprint(buf, fset, f); err != nil {
		return nil, fmt.Errorf("failed to print: %w", err)
	}

	return buf.Bytes(), nil
}

type Xtrace struct {
	Config
	fset *token.FileSet
	src  []byte

	funcByBody   map[ast.Stmt]*FuncInfo
	forByBody    map[ast.Stmt]*ForInfo
	caseByBody   map[ast.Stmt]*CaseInfo
	ifElseByBody map[ast.Stmt]*IfElseInfo

	libraryRequired bool
}

func (s *Xtrace) fragment(pos, end token.Pos) string {
	return string(s.src[pos-1 : end-1])
}
func (s *Xtrace) fragmentLine(pos token.Pos) string {
	begin := pos - 1
	for ; begin > 0; begin-- {
		if s.src[begin-1] == '\n' || s.src[begin-1] == '\r' {
			break
		}
	}
	end := pos
	for ; end < token.Pos(len(s.src)); end++ {
		if s.src[end] == '\n' || s.src[end] == '\r' {
			break
		}
	}
	frag := s.fragment(begin+1, end+1)
	frag, _, _ = strings.Cut(frag, "\n")
	frag, _, _ = strings.Cut(frag, "\r")
	return frag
}

func (s *Xtrace) IdentShowTimestamp() string {
	if s.ShowTimestamp {
		return "true"
	}
	return "false"
}

func (s *Xtrace) IdentShowGoroutine() string {
	if s.ShowGoroutine {
		return "true"
	}
	return "false"
}

type FuncInfo struct {
	Body     *ast.BlockStmt
	FuncDecl *ast.FuncDecl
}

func (i FuncInfo) Signature() (begin, end token.Pos) {
	return i.FuncDecl.Pos(), i.FuncDecl.Body.Pos()
}

func CollectFuncInfo(f *ast.File) (funcByBody map[ast.Stmt]*FuncInfo) {
	funcByBody = map[ast.Stmt]*FuncInfo{}
	ast.PreorderStack(f, nil, func(n ast.Node, s []ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil {
				funcByBody[node.Body] = &FuncInfo{
					Body:     node.Body,
					FuncDecl: node,
				}
			}
		}
		return true
	})
	return funcByBody
}

type ForInfo struct {
	Body  *ast.BlockStmt
	For   *ast.ForStmt
	Range *ast.RangeStmt
}

func (i ForInfo) Variables() (vars []*ast.Ident) {
	if r := i.Range; r != nil {
		if r.Key != nil {
			if ident, ok := r.Key.(*ast.Ident); ok && ident.Name != "_" {
				vars = append(vars, ident)
			}
		}
		if r.Value != nil {
			if ident, ok := r.Value.(*ast.Ident); ok && ident.Name != "_" {
				vars = append(vars, ident)
			}
		}
	}
	if f := i.For; f != nil {
		if assign, ok := f.Init.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
					vars = append(vars, ident)
				}
			}
		}
	}
	return vars
}

func CollectForInfo(f *ast.File) (forByBody map[ast.Stmt]*ForInfo) {
	forByBody = map[ast.Stmt]*ForInfo{}
	ast.PreorderStack(f, nil, func(n ast.Node, s []ast.Node) bool {
		switch node := n.(type) {
		case *ast.ForStmt:
			if node.Body != nil {
				forByBody[node.Body] = &ForInfo{
					Body: node.Body,
					For:  node,
				}
			}
		case *ast.RangeStmt:
			if node.Body != nil {
				forByBody[node.Body] = &ForInfo{
					Body:  node.Body,
					Range: node,
				}
			}
		}
		return true
	})
	return forByBody
}

type CaseInfo struct {
	Case *ast.CaseClause
	Comm *ast.CommClause
}

func (i CaseInfo) CaseLabel() (begin, end token.Pos) {
	if c := i.Case; c != nil {
		return c.Pos(), c.Colon
	}
	if c := i.Comm; c != nil {
		return c.Pos(), c.Colon
	}
	panic("CaseInfo must consist of one of *ast.CaseClause or *ast.CommClause.")
}

func CollectCaseInfo(f *ast.File) (caseByBody map[ast.Stmt]*CaseInfo) {
	caseByBody = map[ast.Stmt]*CaseInfo{}
	ast.PreorderStack(f, nil, func(n ast.Node, s []ast.Node) bool {
		switch node := n.(type) {
		case *ast.CaseClause:
			caseByBody[node] = &CaseInfo{Case: node}
		case *ast.CommClause:
			caseByBody[node] = &CaseInfo{Comm: node}
		}
		return true
	})
	return caseByBody
}

type IfElseInfo struct {
	Body     *ast.BlockStmt
	ElseBody *ast.BlockStmt
	Parents  []*ast.IfStmt
	IfStmt   *ast.IfStmt
}

func (i IfElseInfo) Variables() (vars []*ast.Ident) {
	for _, ifStmt := range append(append([]*ast.IfStmt{}, i.Parents...), i.IfStmt) {
		if ifStmt.Init != nil {
			if assign, ok := ifStmt.Init.(*ast.AssignStmt); ok {
				for _, lhs := range assign.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
						vars = append(vars, ident)
					}
				}
			}
		}
	}
	return vars
}

func CollectIfElseInfo(f *ast.File) (ifElseByBody map[ast.Stmt]*IfElseInfo) {
	ifElseByBody = map[ast.Stmt]*IfElseInfo{}
	ast.PreorderStack(f, nil, func(n ast.Node, s []ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			stack := []*ast.IfStmt{}
			for i := len(s) - 1; i >= 0; i-- {
				stmt, ok := s[i].(*ast.IfStmt)
				if !ok {
					break
				}
				stack = append(stack, stmt)
			}
			mutable.Reverse(stack)

			ifElseByBody[node.Body] = &IfElseInfo{
				Body:    node.Body,
				IfStmt:  node,
				Parents: stack,
			}
			if node.Else != nil {
				if blockBody, ok := node.Else.(*ast.BlockStmt); ok {
					ifElseByBody[blockBody] = &IfElseInfo{
						ElseBody: blockBody,
						IfStmt:   node,
						Parents:  stack,
					}
				}
			}
		}
		return true
	})
	return ifElseByBody
}
