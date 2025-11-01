package goxe

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func ProcessCode(src []byte) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		log.Fatal(err)
	}

	ast.PreorderStack(f, nil, func(n ast.Node, stack []ast.Node) bool {
		// (A) 現在のノードが *ast.Ident (識別子) でなければ、
		// 子ノードの走査を続ける (true)
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		// (B) スタックを逆順 (深い方から浅い方へ) に調べる
		inMainFunc := false
		for i := len(stack) - 1; i >= 0; i-- {
			// スタック内のノードが *ast.FuncDecl (関数宣言) かチェック
			if fn, ok := stack[i].(*ast.FuncDecl); ok {
				// 関数名が 'main' ならフラグを立てる
				if fn.Name.Name == "main" {
					inMainFunc = true
				}
				// 一番近い関数スコープを見つけたらループを抜ける
				break
			}
		}

		// (C) 'main' 関数内にいた場合のみ、識別子名を出力
		if inMainFunc {
			fmt.Printf("Found: %s (at %s)\n", ident.Name, fset.Position(ident.Pos()))
		}

		// Identノードは子を持たないので true/false どちらでも大差ないが、
		// 一般的には子ノードの走査を続けるため true を返す
		return true
	})

}
