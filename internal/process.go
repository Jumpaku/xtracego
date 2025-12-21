package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/samber/lo/mutable"
	"golang.org/x/tools/go/ast/astutil"
)

func ProcessCode(config Config, filename string, src []byte) (dst []byte, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	x := &Xtrace{
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
					x.logFileStatement(c, node)
					x.logFileVariable(c, node)
				}
			case token.CONST:
				x.logFileStatement(c, node)
				x.logFileVariable(c, node)
			}
		case *ast.FuncLit, *ast.FuncDecl:
			var results *ast.FieldList
			switch node := node.(type) {
			case *ast.FuncLit:
				results = node.Type.Results
			case *ast.FuncDecl:
				results = node.Type.Results
			}
			if results != nil {
				count := 0
				for _, param := range results.List {
					if len(param.Names) == 0 {
						count++
						param.Names = []*ast.Ident{ast.NewIdent(fmt.Sprintf("return_%d_%s", count, x.UniqueString))}
					} else {
						for _, name := range param.Names {
							count++
							if name.Name == "_" {
								name.Name = fmt.Sprintf("return_%d_%s", count, x.UniqueString)
							}
						}
					}
				}
				c.Replace(node)
			}
		case ast.Stmt:
			{
				if info, ok := x.funcByBody[node]; ok {
					x.logCallVariables(c, info)
					x.logReturnVariables(c, info)
					x.logCall(c, info)
				}
				if info, ok := x.forByBody[node]; ok {
					x.logForVariables(c, info)
				}
				if info, ok := x.caseByBody[node]; ok {
					x.logCaseStatement(c, info)
				}
				if info, ok := x.ifElseByBody[node]; ok {
					x.logIfVariables(c, info)
					x.logIfElseStatement(c, info)
				}
			}

			x.tryLogLocalStatement(c, node)

			switch node := node.(type) {
			case *ast.DeclStmt:
				if decl, ok := node.Decl.(*ast.GenDecl); ok && decl.Tok == token.VAR {
					x.logLocalVariable(c, node)
				}
			case *ast.AssignStmt:
				if _, ok := c.Parent().(*ast.BlockStmt); ok {
					if node.Tok == token.ASSIGN || node.Tok == token.DEFINE {
						x.logLocalAssignment(c, node)
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

	if x.libraryRequired && x.ResolveType != ResolveType_CommandLineArguments {
		astutil.AddImport(fset, f, config.LibraryImportPath())
	}

	buf := bytes.NewBuffer(nil)
	if err := printer.Fprint(buf, fset, f); err != nil {
		return nil, fmt.Errorf("failed to print: %w", err)
	}

	return buf.Bytes(), nil
}

type FuncInfo struct {
	Body     *ast.BlockStmt
	FuncDecl *ast.FuncDecl
	FuncLit  *ast.FuncLit
}

func (i FuncInfo) Signature() (begin, end token.Pos) {
	if i.FuncDecl != nil {
		return i.FuncDecl.Pos(), i.FuncDecl.Body.Pos()
	}
	if i.FuncLit != nil {
		return i.FuncLit.Pos(), i.FuncLit.Body.Pos()
	}
	panic("FuncInfo must consist of one of *ast.FuncDecl or *ast.FuncLit.")
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
		case *ast.FuncLit:
			if node.Body != nil {
				funcByBody[node.Body] = &FuncInfo{
					Body:    node.Body,
					FuncLit: node,
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
