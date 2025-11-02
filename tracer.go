package xtracego

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/samber/lo/mutable"
)

type Xtrace struct {
	fset      *token.FileSet
	src       []byte
	lineWidth int
	prefix    string

	funcByBody   map[ast.Stmt]*FuncInfo
	forByBody    map[ast.Stmt]*ForInfo
	caseByBody   map[ast.Stmt]*CaseInfo
	ifElseByBody map[ast.Stmt]*IfElseInfo
}

func (s Xtrace) fragment(pos, end token.Pos) string {
	return string(s.src[pos-1 : end-1])
}
func (s Xtrace) fragmentLine(pos token.Pos) string {
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
			if node.Body != nil && len(node.Body.List) > 0 {
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
