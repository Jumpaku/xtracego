package main

const _ = iota

const (
	a1 = iota
	a2
)

const (
	b1, b2 = iota, iota
)

var _ = 1
var x = 1
var (
	y1 = 1
	y2 = 1
)

var (
	z1, z2 = 1, 1
)

func sample() {
	v := 1
	var u = 2
	u, v = v, u
}
