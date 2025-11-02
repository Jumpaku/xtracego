# xtracego

xtracego is a command-line tool to run Go source code injecting xtrace.

for rewriting Go source code to inject tracing and logging for function calls, statements, variables, and more.
It provides a command-line interface to rewrite, build, and run Go programs with tracing enabled, making it easier to debug and understand code execution.

## Features

### Xtrace

- Trace basic statements
- Trace function and method calls
- Trace switch/select cases
- Trace variables and constants

## Installation

```sh
go install github.com/jumpaku/xtracego/cmd/xtracego@latest
```

## Usage

### Rewrite source files to inject xtrace

```sh
xtracego rewrite -o=out_dir ./path/to/package
```

### Build source files with injected xtrace

```sh
xtracego build -o=build_dir ./path/to/package
```

### Run source files with injected xtrace

```sh
xtracego run ./path/to/package
```

## Example

### FizzBuzz

Source code:

```go
// examples/fizzbuzz/main.go
package main

import "fmt"

const N = 20

func main() {
	for i := 1; i <= N; i++ {
		if i%15 == 0 {
			fmt.Println("FizzBuzz")
		} else if i%3 == 0 {
			fmt.Println("Fizz")
		} else if i%5 == 0 {
			fmt.Println("Buzz")
		} else {
			fmt.Println(i)
		}
	}
}
```

Run:

```sh
xtracego run ./examples/fizzbuzz
```

Got trace output from stderr:

```
2025/11/03 02:16:59 const N = 20 ........................ [ /path/to/examples/fizzbuzz/main.go:5:7 ]
2025/11/03 02:16:59 [VAR] N=20
2025/11/03 02:16:59 [CALL] func main() 
2025/11/03 02:16:59     for i := 1; i <= N; i++ { ....... [ /path/to/examples/fizzbuzz/main.go:8:2 ]
2025/11/03 02:16:59 [VAR] i=1
2025/11/03 02:16:59         if i%15 == 0 { .............. [ /path/to/examples/fizzbuzz/main.go:9:3 ]
2025/11/03 02:16:59         } else if i%3 == 0 { ........ [ /path/to/examples/fizzbuzz/main.go:11:10 ]
2025/11/03 02:16:59         } else if i%5 == 0 { ........ [ /path/to/examples/fizzbuzz/main.go:13:10 ]
2025/11/03 02:16:59         } else { .................... [ /path/to/examples/fizzbuzz/main.go:15:3 ]
2025/11/03 02:16:59             fmt.Println(i) .......... [ /path/to/examples/fizzbuzz/main.go:16:4 ]
2025/11/03 02:16:59 [VAR] i=2
2025/11/03 02:16:59         if i%15 == 0 { .............. [ /path/to/examples/fizzbuzz/main.go:9:3 ]
2025/11/03 02:16:59         } else if i%3 == 0 { ........ [ /path/to/examples/fizzbuzz/main.go:11:10 ]
2025/11/03 02:16:59         } else if i%5 == 0 { ........ [ /path/to/examples/fizzbuzz/main.go:13:10 ]
2025/11/03 02:16:59         } else { .................... [ /path/to/examples/fizzbuzz/main.go:15:3 ]
2025/11/03 02:16:59             fmt.Println(i) .......... [ /path/to/examples/fizzbuzz/main.go:16:4 ]
2025/11/03 02:16:59 [VAR] i=3
2025/11/03 02:16:59         if i%15 == 0 { .............. [ /path/to/examples/fizzbuzz/main.go:9:3 ]
2025/11/03 02:16:59         } else if i%3 == 0 { ........ [ /path/to/examples/fizzbuzz/main.go:11:10 ]
2025/11/03 02:16:59             fmt.Println("Fizz") ..... [ /path/to/examples/fizzbuzz/main.go:12:4 ]
...
```


Got output from stdout:

```
1
2
Fizz
...
```


## Documentation

### Command-line interface

See the following CLI documentation.

https://github.com/Jumpaku/xtracego/blob/main/docs/xtracego.md
