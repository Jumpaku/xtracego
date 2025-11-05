# xtracego


## xtracego

### Syntax

```shell
xtracego [<option>]...
```

### Options

* `-copy-only=<string>`  (default=`""`):  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`""`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

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

* `-copy-only=<string>`  (default=`""`):  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`""`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-go-build-arg=<string>`, `-a=<string>`  (default=`""`):  
  Arguments to be passed to the go build command.  
  If there are multiple arguments for go build, this option can be specified multiple times.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

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
  Path to a local directory of the main package to be rewritten.  




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

* `-copy-only=<string>`  (default=`""`):  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`""`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

* `-output-directory=<string>`, `-o=<string>`  (default=`""`):  
  Output directory to place the rewritten source files of the package.  
  This option is required.  

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
  Path to a local directory of the main package to be rewritten.  




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

* `-copy-only=<string>`  (default=`""`):  
  Specifies source files not to be rewritten but only copied by regular expressions.  
  If a source file is included in the package and its absolute path matches this regular expression, it is only copied to the output directory.  

* `-copy-only-not=<string>`  (default=`""`):  
  Same as -copy-only but source files whose absolute path  **DO NOT MATCH**  this regular expression are only copied.  

* `-go-build-arg=<string>`, `-a=<string>`  (default=`""`):  
  Arguments to be passed to the go run command.  
  If there are multiple arguments for go build, this option can be specified multiple times.  

* `-goroutine[=<boolean>]`  (default=`true`),  
  `-no-goroutine[=<boolean>]`:  
  Whether show goroutine ID or not.  

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
  Path to a local directory of the main package to be rewritten, followed by arguments to be passed to the main function.  

1. `[<arguments:string>]...`  
  Arguments to be passed to the main function.  




