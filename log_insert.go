package goxe

import (
	"go/token"
	"strings"
)

type LogInsert struct {
	fset      *token.FileSet
	src       []byte
	lineWidth int
	prefix    string
}

func (s LogInsert) fragment(pos, end token.Pos) string {
	return string(s.src[pos-1 : end-1])
}
func (s LogInsert) fragmentLine(pos, end token.Pos) string {
	begin := pos - 1
	for ; begin > 0; begin-- {
		if s.src[begin-1] == '\n' || s.src[begin-1] == '\r' {
			break
		}
	}
	frag := s.fragment(begin+1, end)
	frag, _, _ = strings.Cut(frag, "\n")
	frag, _, _ = strings.Cut(frag, "\r")
	return frag
}
