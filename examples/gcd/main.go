//go:build ignore

// examples/gcd/main.go
package main

import "fmt"

func main() {
	var x, y int64 = 664, 576
	fmt.Println(gcd(x, y))
}

func gcd(a, b int64) int64 {
	switch {
	case b == 0:
		return a
	default:
		return gcd(b, a%b)
	}
}
