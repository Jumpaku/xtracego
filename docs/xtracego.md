# xtracego (v0.0.8)


## xtracego

### Syntax

```shell
xtracego [<option>]...
```

### Options

* `-copy-only=<string> ...`  :  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`".*"`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

* `-help[=<boolean>]`, `-h[=<boolean>]`  (default=`false`):  
  Prints help message.  

* `-seed=<integer>`  (default=`0`):  
  Random seed for reproducibility of rewritten source files.  
  If not specified, the seed is generated randomly.  

* `-timestamp[=<boolean>]`  (default=`true`),  
  `-no-timestamp[=<boolean>]`:  
  Whether show timestamp or not.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling and returning functions and methods or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  

### Subcommands

* build:  
  Rewrites the source files in the specified package and places these files in the build directory.  
  Executes go build at the specified directory with the given arguments.  

* rewrite:  
  Rewrites the source files in the specified package and places these files in the output directory.  
  The rewritten files includes Go code to log trace information.  
  If go.mod of the module of the package is found, it is copied to the output directory.  

* run:  
  Rewrites the source files in the specified package and places these files in a temporary directory.  
  Executes go build at the temporary directory with the given arguments.  
  Thereafter, the built executable file is executed at the current working directory.  

* version:  
  Prints the version of xtracego.  




## xtracego build

### Description

Rewrites the source files in the specified package and places these files in the build directory.
Executes go build at the specified directory with the given arguments.

### Syntax

```shell
xtracego build [<option>|<argument>]... [-- [<argument>]...]
```

### Options

* `-build-directory=<string>`, `-o=<string>`  (default=`""`):  
  The source files included the specified package are rewritten and placed in this directory which is used as a current working directory to execute go build.  
  This option is required.  

* `-copy-only=<string> ...`  :  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`".*"`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-go-build-arg=<string> ...`, `-a=<string> ...`  :  
  Arguments to be passed to the go build command.  
  If there are multiple arguments for go build, this option can be specified multiple times.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

* `-help[=<boolean>]`, `-h[=<boolean>]`  (default=`false`):  
  Prints help message.  

* `-seed=<integer>`  (default=`0`):  
  Random seed for reproducibility of rewritten source files.  
  If not specified, the seed is generated randomly.  

* `-timestamp[=<boolean>]`  (default=`true`),  
  `-no-timestamp[=<boolean>]`:  
  Whether show timestamp or not.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling and returning functions and methods or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  

### Arguments

0. `<package:string>`  
  Package to be rewritten and built.  
  The way to specify the package is as same as xtracego rewrite command.  




## xtracego rewrite

### Description

Rewrites the source files in the specified package and places these files in the output directory.
The rewritten files includes Go code to log trace information.
If go.mod of the module of the package is found, it is copied to the output directory.

### Syntax

```shell
xtracego rewrite [<option>|<argument>]... [-- [<argument>]...]
```

### Options

* `-copy-only=<string> ...`  :  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`".*"`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

* `-help[=<boolean>]`, `-h[=<boolean>]`  (default=`false`):  
  Prints help message.  

* `-output-directory=<string>`, `-o=<string>`  (default=`""`):  
  Output directory to place the rewritten source files of the package.  
  This option is required.  

* `-seed=<integer>`  (default=`0`):  
  Random seed for reproducibility of rewritten source files.  
  If not specified, the seed is generated randomly.  

* `-timestamp[=<boolean>]`  (default=`true`),  
  `-no-timestamp[=<boolean>]`:  
  Whether show timestamp or not.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling and returning functions and methods or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  

### Arguments

0. `<package:string>`  
  Package to be rewritten can be specified by path to a local directory or paths to local source files of a main package.  
    
  When the package is specified with a local directory path, go.mod must be found at the ancestors of the current working directory.  
  Dependencies in the same module and external dependencies are resolved via the go.mod.  
    
  When the package is specified with local source file paths, the source files must have extension .go, be in the same directory, be in the main package, and contain only one main function.  
  If go.mod is found at the ancestors of the current working directory, dependencies in the same module and external dependencies are resolved via the go.mod.  




## xtracego run

### Description

Rewrites the source files in the specified package and places these files in a temporary directory.
Executes go build at the temporary directory with the given arguments.
Thereafter, the built executable file is executed at the current working directory.

### Syntax

```shell
xtracego run [<option>|<argument>]... [-- [<argument>]...]
```

### Options

* `-copy-only=<string> ...`  :  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`".*"`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-go-build-arg=<string> ...`, `-a=<string> ...`  :  
  Arguments to be passed to the go run command.  
  If there are multiple arguments for go build, this option can be specified multiple times.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

* `-help[=<boolean>]`, `-h[=<boolean>]`  (default=`false`):  
  Prints help message.  

* `-seed=<integer>`  (default=`0`):  
  Random seed for reproducibility of rewritten source files.  
  If not specified, the seed is generated randomly.  

* `-timestamp[=<boolean>]`  (default=`true`),  
  `-no-timestamp[=<boolean>]`:  
  Whether show timestamp or not.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling and returning functions and methods or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  

* `-width=<integer>`, `-w=<integer>`  (default=`0`):  
  Terminal width to be used for formatting trace messages.  

### Arguments

0. `<package:string>`  
  Package to be rewritten and built.  
  The way to specify the package is as same as xtracego rewrite command.  

1. `[<arguments:string>]...`  
  Arguments to be passed to the main function.  




## xtracego version

### Description

Prints the version of xtracego.

### Syntax

```shell
xtracego version [<option>]...
```

### Options

* `-copy-only=<string> ...`  :  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`".*"`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

* `-help[=<boolean>]`, `-h[=<boolean>]`  (default=`false`):  
  Prints help message.  

* `-seed=<integer>`  (default=`0`):  
  Random seed for reproducibility of rewritten source files.  
  If not specified, the seed is generated randomly.  

* `-timestamp[=<boolean>]`  (default=`true`),  
  `-no-timestamp[=<boolean>]`:  
  Whether show timestamp or not.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling and returning functions and methods or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  




