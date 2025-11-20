# xtracego

xtracego is a command-line tool designed to automatically inject execution traces (xtrace) into Go source code.
By manipulating the Abstract Syntax Tree (AST), it enables to log statements, function calls, and variable states to stderr.
Its goal is to enable easy debugging and logging for Go scripting, similar to the `set -x` functionality in shell scripts.

## Features

- Run: Execute Go source files directly with injected traces.
- Build: Compile source files into an executable with injected traces.
- Rewrite: Generate modified source files with injected traces.

The following execution trace information is available:

- Execution of basic statements
- Function and method calls
- Values of variables and constants

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
2025-11-03T11:55:22Z [ 1] :const N = 20 ............................. [ /path/to/examples/fizzbuzz/main.go:6:7 ]
2025-11-03T11:55:22Z [ 1] :[VAR] N=20
2025-11-03T11:55:22Z [ 1] :[CALL] (main.main)
2025-11-03T11:55:22Z [ 1] :    for i := 1; i <= N; i++ { ............ [ /path/to/examples/fizzbuzz/main.go:9:2 ]
2025-11-03T11:55:22Z [ 1] :[VAR] i=1
2025-11-03T11:55:22Z [ 1] :        if i%15 == 0 { ................... [ /path/to/examples/fizzbuzz/main.go:10:3 ]
2025-11-03T11:55:22Z [ 1] :        } else if i%3 == 0 { ............. [ /path/to/examples/fizzbuzz/main.go:12:10 ]
2025-11-03T11:55:22Z [ 1] :        } else if i%5 == 0 { ............. [ /path/to/examples/fizzbuzz/main.go:14:10 ]
2025-11-03T11:55:22Z [ 1] :        } else { ......................... [ /path/to/examples/fizzbuzz/main.go:16:3 ]
2025-11-03T11:55:22Z [ 1] :            fmt.Println(i) ............... [ /path/to/examples/fizzbuzz/main.go:17:4 ]
2025-11-03T11:55:22Z [ 1] :[VAR] i=2
2025-11-03T11:55:22Z [ 1] :        if i%15 == 0 { ................... [ /path/to/examples/fizzbuzz/main.go:10:3 ]
2025-11-03T11:55:22Z [ 1] :        } else if i%3 == 0 { ............. [ /path/to/examples/fizzbuzz/main.go:12:10 ]
2025-11-03T11:55:22Z [ 1] :        } else if i%5 == 0 { ............. [ /path/to/examples/fizzbuzz/main.go:14:10 ]
2025-11-03T11:55:22Z [ 1] :        } else { ......................... [ /path/to/examples/fizzbuzz/main.go:16:3 ]
2025-11-03T11:55:22Z [ 1] :            fmt.Println(i) ............... [ /path/to/examples/fizzbuzz/main.go:17:4 ]
2025-11-03T11:55:22Z [ 1] :[VAR] i=3
2025-11-03T11:55:22Z [ 1] :        if i%15 == 0 { ................... [ /path/to/examples/fizzbuzz/main.go:10:3 ]
2025-11-03T11:55:22Z [ 1] :        } else if i%3 == 0 { ............. [ /path/to/examples/fizzbuzz/main.go:12:10 ]
2025-11-03T11:55:22Z [ 1] :            fmt.Println("Fizz") .......... [ /path/to/examples/fizzbuzz/main.go:13:4 ]
...
```


Got output from stdout:

```
1
2
Fizz
...
```

Examples are available in `examples` ( https://github.com/Jumpaku/xtracego/tree/main/examples ).

## Motivation

The goal of xtracego is to facilitate easy debugging and logging for Go scripts by providing execution traces similar to `set -x` in shell scripts.
Shell scripts offer the convenient `set -x` that prints each command to stderr before execution, which is useful for debugging and logging.
For example, if you have the following shell script:
```shell
#!/bin/sh
set -x
pwd
echo "Hello, world!"
```
executing `sh setx.sh` outputs:
```
# +pwd
# /path/to/working/directory
# +echo "Hello, world!"
# Hello, world!
```

When it comes to Go scripting, the conventional debugging or logging techniques have the following drawbacks:
- `log.Println`: Simple but tedious to insert manually everywhere.
- Stacktrace: Shows where a panic occurred but lacks execution flow and variable history.
- Debuggers: Powerful for interactive debugging but not suitable for logging execution history.

To overcome these drawbacks, xtracego automatically injects execution traces by modifying the abstract syntax tree (AST) of the source files.
Therefore, xtracego makes debugging and logging in Go scripting as easy as using `set -x`.

## Installation

### Using Go install

```shell
go install github.com/Jumpaku/xtracego/cmd/xtracego@latest
```

### Using Docker

```shell
docker run -i -v $(pwd):/workspace ghcr.io/jumpaku/xtracego:latest xtracego
```

### Downloading executable binary files

https://github.com/Jumpaku/xtracego/releases

Note that the downloaded executable binary file may require a security confirmation before it can be run.

### Building from source

```shell
git clone https://github.com/Jumpaku/xtracego.git
cd xtracego
go install ./cmd/xtracego
```

## Usage

### Run Go source files with xtrace

```sh
xtracego run ./path/to/package
```

### Build an executable file from source files with xtrace

```sh
xtracego build -o=build_dir ./path/to/package
```

### Generate source files injected xtrace

```sh
xtracego rewrite -o=out_dir ./path/to/package
```

## Documentation

### Command-line interface

For detailed CLI usage, please see:

https://github.com/Jumpaku/xtracego/blob/main/docs/xtracego.md

## Limitation

- Comments are not preserved during the AST rewriting process. 
  Therefore, compiler directives (e.g., `//go:embed`) are ignored, which may affect builds relying on them.
