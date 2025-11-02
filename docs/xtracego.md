# xtracego


## xtracego

### Syntax

```shell
xtracego [<option>]...
```

### Options

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling functions and methods or not.  

* `-trace-case[=<boolean>]`  (default=`true`),  
  `-no-trace-case[=<boolean>]`:  
  Whether trace cases of switch and select statements or not.  

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

* `-go-build-arg=<string>`, `-a=<string>`  (default=`""`):  
  Arguments to be passed to the go build command.  
  If there are multiple arguments for go build, this option can be specified multiple times.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling functions and methods or not.  

* `-trace-case[=<boolean>]`  (default=`true`),  
  `-no-trace-case[=<boolean>]`:  
  Whether trace cases of switch and select statements or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  

### Arguments

0. `[<package:string>]...`  
  Path to a local directory of the main package or paths to source files in the package to be rewritten.  




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

* `-output-directory=<string>`, `-o=<string>`  (default=`""`):  
  Output directory to place the rewritten source files of the package.  
  This option is required.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling functions and methods or not.  

* `-trace-case[=<boolean>]`  (default=`true`),  
  `-no-trace-case[=<boolean>]`:  
  Whether trace cases of switch and select statements or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  

### Arguments

0. `[<package:string>]...`  
  Path to a local directory of the main package or paths to source files in the package to be rewritten.  




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

* `-go-build-arg=<string>`, `-a=<string>`  (default=`""`):  
  Arguments to be passed to the go run command.  
  If there are multiple arguments for go build, this option can be specified multiple times.  

* `-trace-call[=<boolean>]`  (default=`true`),  
  `-no-trace-call[=<boolean>]`:  
  Whether trace calling functions and methods or not.  

* `-trace-case[=<boolean>]`  (default=`true`),  
  `-no-trace-case[=<boolean>]`:  
  Whether trace cases of switch and select statements or not.  

* `-trace-stmt[=<boolean>]`  (default=`true`),  
  `-no-trace-stmt[=<boolean>]`:  
  Whether trace basic statements or not.  

* `-trace-var[=<boolean>]`  (default=`true`),  
  `-no-trace-var[=<boolean>]`:  
  Whether trace variables and constants or not.  

* `-verbose[=<boolean>]`, `-v[=<boolean>]`  (default=`false`):  
  Whether to output verbose messages or not.  

### Arguments

0. `[<package_and_arguments:string>]...`  
  Path to a local directory of the main package or paths to source files in the package to be rewritten, followed by arguments to be passed to the main function.  
  If a local directory is given as the first argument, the rest of the arguments is treated as arguments to the main function.  
  If source files, each of which ends with '.go', are given as the first arguments, the rest of the arguments is treated as arguments to the main function.  
  Arguments after the first '--' are treated as arguments to the main function.  
  * Example 1:  
    xtracego run /path/to/main/package arg1 arg2 --> package=/path/to/main/package, arguments=[arg1, arg2]  
  * Example 2:  
    xtracego run ./path/to/main/package arg1 arg2 --> package=./path/to/main/package, arguments=[arg1, arg2]  
  * Example 3:  
    xtracego run ./path/to/main/package -- arg1 arg2 --> package=./path/to/main/package, arguments=[arg1, arg2]  
  * Example 4:  
    xtracego run ./source.go ./files.go ./arg.go arg1 arg2 --> package=[./source.go, ./files.go, ./arg.go], arguments=[arg1, arg2]   
  * Example 5:  
    xtracego run ./source.go ./files.go -- ./arg.go arg1 arg2 --> package=[./source.go, ./files.go], arguments=[./arg.go, arg1, arg2]   




