package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Jumpaku/xtracego"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

//go:generate cyamli generate golang -schema-path=cli.cyamli.yaml -out-path=cli.gen.go
//go:generate cyamli generate docs -format=markdown -schema-path=cli.cyamli.yaml -out-path=../../docs/xtracego.md
func main() {
	if err := Run(&cliHandler{}, os.Args); err != nil {
		log.Panicf("error: %+v\n", err)
	}
}

func requireOption[T int64 | bool | string](subcommand []string, option string, value T) T {
	var zero T
	panicIf(value == zero, "option %s is required\n%s", option, GetDoc(subcommand))
	return value
}

func panicIfError(err error, format string, args ...any) {
	if err != nil {
		log.Panicln(fmt.Sprintf(format, args...) + fmt.Sprintf(": %+v", err))
	}
}

func panicIf(shouldPanic bool, format string, args ...any) {
	if shouldPanic {
		log.Panicln(fmt.Sprintf(format, args...))
	}
}

type cliHandler struct {
	verbose bool
}

var _ CLIHandler = &cliHandler{}

func (h *cliHandler) logf(format string, args ...any) {
	if h.verbose {
		log.Printf(format+"\n", args...)
	}
}

func (h *cliHandler) Run_Version(_ Input_Version) error {
	fmt.Println(GetVersion())
	return nil
}

func (h *cliHandler) Run(input Input) error {
	log.Println(GetDoc(input.Subcommand))
	return nil
}

func (h *cliHandler) Run_Rewrite(input Input_Rewrite) (err error) {
	h.verbose = input.Opt_Verbose

	outDir := requireOption(input.Subcommand, "-output-directory", input.Opt_OutputDirectory)

	pkg := h.resolvePackage(input.Arg_Package)

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   pkg.ResolveType,
		ModuleName:    pkg.Module,
		UniqueString:  generateUniqueString(input.Opt_Seed),
	}

	h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot)

	h.saveLibraryFiles(cfg, outDir)

	if pkg.ResolveType == xtracego.ResolveType_CommandLineArguments {
		h.saveGoModFile(cfg, outDir)
	}

	return nil
}

func (h cliHandler) Run_Build(input Input_Build) (err error) {
	h.verbose = input.Opt_Verbose

	outDir := requireOption(input.Subcommand, "-build-directory", input.Opt_BuildDirectory)

	pkg := h.resolvePackage(input.Arg_Package)

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   pkg.ResolveType,
		ModuleName:    pkg.Module,
		UniqueString:  generateUniqueString(input.Opt_Seed),
	}

	h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot)

	h.saveLibraryFiles(cfg, outDir)

	packageDir := h.getBuildPackageDir(pkg, outDir)
	if pkg.ResolveType == xtracego.ResolveType_CommandLineArguments {
		h.saveGoModFile(cfg, outDir)
	}

	h.execGoModTidy(outDir)

	h.execGoBuild(input.Opt_GoBuildArg, packageDir, outDir)

	return nil
}

func (h cliHandler) Run_Run(input Input_Run) error {
	h.verbose = input.Opt_Verbose

	outDir, err := os.MkdirTemp("", "xtracego_*")
	panicIfError(err, "failed to create temp dir")
	defer os.RemoveAll(outDir)

	pkg := h.resolvePackage(input.Arg_Package)

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   pkg.ResolveType,
		ModuleName:    pkg.Module,
		UniqueString:  generateUniqueString(input.Opt_Seed),
	}

	h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot)

	h.saveLibraryFiles(cfg, outDir)

	packageDir := h.getBuildPackageDir(pkg, outDir)
	if pkg.ResolveType == xtracego.ResolveType_CommandLineArguments {
		h.saveGoModFile(cfg, outDir)
	}

	h.execGoModTidy(outDir)

	execFile, err := filepath.Abs(filepath.Join(outDir, cfg.ExecutableFileName()))
	panicIfError(err, "failed to get absolute path")

	h.execGoBuild(append(input.Opt_GoBuildArg, "-o", execFile), packageDir, outDir)

	h.execBuiltFile(input, execFile)

	return nil
}

func generateUniqueString(seed int64) string {
	if seed == 0 {
		seed = time.Now().Unix()
	}
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	r := rand.New(rand.NewSource(seed))
	v := []byte{}
	for i := 0; i < 8; i++ {
		v = append(v, alphabet[r.Intn(len(alphabet))])
	}
	return string(v)
}

func (h cliHandler) resolvePackage(packageArg string) xtracego.ResolvedPackage {
	pkg, err := xtracego.ResolvePackage(packageArg)
	panicIfError(err, "failed to resolve package")
	return pkg
}

func (h *cliHandler) transformSourceFiles(
	cfg xtracego.Config,
	pkg xtracego.ResolvedPackage,
	outDir string,
	copyOnlyRegexpStr []string,
	copyOnlyNotRegexpStr string,
) {
	srcDir, sourceFiles := pkg.PackageDir, append([]string{}, pkg.SourceFiles...)
	if pkg.ResolveType != xtracego.ResolveType_CommandLineArguments {
		srcDir, sourceFiles = filepath.Dir(pkg.GoModFile), append(sourceFiles, pkg.GoModFile)
	}

	copyOnlyRegexp := []*regexp.Regexp{}
	for _, s := range copyOnlyRegexpStr {
		re, err := regexp.Compile(s)
		panicIfError(err, "failed to compile regexp '%s'", s)

		copyOnlyRegexp = append(copyOnlyRegexp, re)
	}
	copyOnlyNotRegexp, err := regexp.Compile(copyOnlyNotRegexpStr)
	panicIfError(err, "failed to compile regexp '%s'", copyOnlyNotRegexpStr)

	eg, _ := errgroup.WithContext(context.Background())
	for _, srcFile := range sourceFiles {
		eg.Go(func() error {
			isCopyOnly := (!copyOnlyNotRegexp.MatchString(srcFile)) ||
				lo.SomeBy(copyOnlyRegexp, func(r *regexp.Regexp) bool { return r.MatchString(srcFile) })
			isGoSource := strings.HasSuffix(srcFile, ".go")

			relToFile, err := filepath.Rel(srcDir, srcFile)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			dstFile := filepath.Join(outDir, relToFile)

			err = xtracego.TransformFile(srcFile, dstFile, func(r io.Reader, w io.Writer) (err error) {
				if isGoSource && !isCopyOnly {
					h.logf("[rewrite] %s -> %s", srcFile, dstFile)
					src := bytes.NewBuffer(nil)
					if _, err := io.Copy(src, r); err != nil {
						return fmt.Errorf("failed to copy file: %w", err)
					}
					buf, err := xtracego.ProcessCode(cfg, srcFile, src.Bytes())
					if err != nil {
						return fmt.Errorf("failed to rewrite file: %w", err)
					}
					if _, err := w.Write(buf); err != nil {
						return fmt.Errorf("failed to write file: %w", err)
					}
				} else {
					h.logf("[copy] %s -> %s", srcFile, dstFile)
					if _, err := io.Copy(w, r); err != nil {
						return fmt.Errorf("failed to copy file: %w", err)
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to rewrite file: %w", err)
			}
			return nil
		})
	}

	err = eg.Wait()
	panicIfError(err, "failed to wait")
}

func (h *cliHandler) saveLibraryFiles(cfg xtracego.Config, outDir string) {
	dst := filepath.Join(outDir, cfg.LibraryFileName())
	if cfg.ResolveType != xtracego.ResolveType_CommandLineArguments {
		dst = filepath.Join(outDir, cfg.LibraryPackageName(), cfg.LibraryFileName())
	}

	buf := bytes.NewBuffer(nil)
	err := xtracego.GetLibraryCode(cfg.LibraryPackageName(), cfg.UniqueString, buf)
	panicIfError(err, "failed to generate library")

	h.logf("[add] %s", dst)
	err = xtracego.SaveFile(dst, buf.String())
	panicIfError(err, "failed to save library")
}

func (h *cliHandler) saveGoModFile(cfg xtracego.Config, outDir string) {
	dst := filepath.Join(outDir, "go.mod")
	panicIf(cfg.ResolveType != xtracego.ResolveType_CommandLineArguments, "go.mod is not required")

	h.logf("[add] %s", dst)
	err := xtracego.SaveFile(dst, fmt.Sprintf(`module %s`, cfg.UniqueString))
	panicIfError(err, "failed to save go.mod file")
}

func (h cliHandler) getBuildPackageDir(pkg xtracego.ResolvedPackage, outDir string) string {
	if pkg.ResolveType == xtracego.ResolveType_CommandLineArguments {
		return "."
	}
	cwd, err := os.Getwd()
	panicIfError(err, "failed to get current directory")
	relToPkg, err := filepath.Rel(cwd, pkg.PackageDir)
	panicIfError(err, "failed to get relative path")
	outDir, err = filepath.Abs(outDir)
	panicIfError(err, "failed to get absolute path")
	return filepath.Join(outDir, relToPkg)
}

func (h cliHandler) execGoModTidy(outDir string) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
	h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
	err := cmd.Run()
	panicIfError(err, "failed to run go mod tidy")
}

func (h cliHandler) execGoBuild(buildArgs []string, buildPackageDir string, outDir string) {
	args := append(append([]string{"build"}, buildArgs...), buildPackageDir)
	cmd := exec.Command("go", args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
	h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
	err := cmd.Run()
	panicIfError(err, "failed to run go build")
}

func (h cliHandler) execBuiltFile(input Input_Run, execFile string) {
	cmd := exec.Command(execFile, input.Arg_Arguments...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	h.logf("[exec] %s", cmd.String())
	err := cmd.Run()
	panicIfError(err, "failed to run built file")
}
