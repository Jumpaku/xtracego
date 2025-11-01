package dir

import (
	"embed"
	_ "embed"
	"fmt"
)

//go:embed embed.txt
var embedTxt string

//go:embed *.txt
var contents embed.FS

func SampleFunc() {
	fmt.Println("dir/func.go")
}
