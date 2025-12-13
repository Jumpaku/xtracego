//go:build ignore

// examples/funclit/main.go
package main

import "fmt"

func main() {
	_ = func(msg string) int64 {
		fmt.Println(msg)
		return 1
	}("Hello, Immediate Execution!")
	func() {}()
	func() int { return 0 }()
	func() (int, int) { return 0, 0 }()
	func() (i int, _ int) { return 0, 0 }()
	func() (i, j int, _, _ int) { return 0, 0, 0, 0 }()
}
