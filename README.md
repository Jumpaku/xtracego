# xtracego

xtracego is a command-line tool to run Go source code injecting xtrace.

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

## Features

- Run Go source files directly with injected xtrace.
- Build an executable file from source files with injected xtrace.
- Rewrite source files to inject xtrace.

### Xtrace

Xtrace is execution trace information. The following traces are available:

- Traces of basic statements
- Traces of function and method calls
- Traces of switch/select cases
- Traces of variables and constants

## Installation

```sh
go install github.com/jumpaku/xtracego/cmd/xtracego@latest
```

## Usage

### Run Go source files directly with injected xtrace

```sh
xtracego run ./path/to/package
```

### Build an executable file from source files with injected xtrace

```sh
xtracego build -o=build_dir ./path/to/package
```

### Rewrite source files to inject xtrace

```sh
xtracego rewrite -o=out_dir ./path/to/package
```

## Documentation

### Command-line interface

See the following CLI documentation:

https://github.com/Jumpaku/xtracego/blob/main/docs/xtracego.md
