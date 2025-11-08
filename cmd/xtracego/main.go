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

//go:generate cyamli generate golang -schema-path=cli.yaml -out-path=cli.gen.go
//go:generate cyamli generate docs -format=markdown -schema-path=cli.yaml -out-path=../../docs/xtracego.md
func main() {
	if err := Run(&cliHandler{}, os.Args); err != nil {
		log.Panicf("error: %+v\n", err)
	}
}

func requireOption[T int64 | bool | string](subcommand []string, option string, value T) T {
	var zero T
	if value == zero {
		log.Panicln(fmt.Sprintf("option %s is required\n", option) + GetDoc(subcommand))
	}
	return value
}

func panicfIfError(err error, format string, args ...any) {
	if err != nil {
		log.Panicln(fmt.Sprintf(format, args...) + fmt.Sprintf(": %+v", err))
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

func (h *cliHandler) Run(input Input) error {
	log.Println(GetDoc(input.Subcommand))
	return nil
}

func (h *cliHandler) Run_Rewrite(input Input_Rewrite) (err error) {
	h.verbose = input.Opt_Verbose

	outDir := requireOption(input.Subcommand, "-output-directory", input.Opt_OutputDirectory)

	resolveType, _, pkg, err := h.resolvePackage(input.Arg_Package)
	panicfIfError(err, "failed to resolve package")

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   resolveType,
		ModulePath:    pkg.Module,
		UniqueString:  generateUniqueString(time.Now().Unix()),
	}

	err = h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot)
	panicfIfError(err, "failed to clone source files")

	_, err = h.saveLibraryFiles(cfg, outDir)
	panicfIfError(err, "failed to save xtracego library")

	return nil
}

func (h cliHandler) Run_Build(input Input_Build) error {
	h.verbose = input.Opt_Verbose

	outDir := requireOption(input.Subcommand, "-build-directory", input.Opt_BuildDirectory)

	resolveType, packageArgs, pkg, err := h.resolvePackage(input.Arg_Package)
	panicfIfError(err, "failed to resolve package")

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   resolveType,
		ModulePath:    pkg.Module,
		UniqueString:  generateUniqueString(time.Now().Unix()),
	}

	err = h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot)
	panicfIfError(err, "failed to clone source files")

	libSource, err := h.saveLibraryFiles(cfg, outDir)
	panicfIfError(err, "failed to save xtracego library")

	if pkg.GoModFile == "" {
		packageArgs = append(packageArgs, libSource)
	} else {
		err := h.execGoModTidy(outDir)
		panicfIfError(err, "failed to run go mod tidy")
	}

	err = h.execGoBuild(input.Opt_GoBuildArg, packageArgs, outDir)
	panicfIfError(err, "failed to run go build")

	return nil
}

func (h cliHandler) Run_Run(input Input_Run) error {
	h.verbose = input.Opt_Verbose

	outDir, err := os.MkdirTemp("", "xtracego_*")
	panicfIfError(err, "failed to create temp dir")
	defer os.RemoveAll(outDir)

	resolveType, packageArgs, pkg, err := h.resolvePackage(input.Arg_Package)
	panicfIfError(err, "failed to resolve package")

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   resolveType,
		ModulePath:    pkg.Module,
		UniqueString:  generateUniqueString(time.Now().Unix()),
	}

	err = h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot)
	panicfIfError(err, "failed to clone source files")

	libSource, err := h.saveLibraryFiles(cfg, outDir)
	panicfIfError(err, "failed to save xtracego library")
	if pkg.GoModFile == "" {
		packageArgs = append(packageArgs, libSource)
	} else {
		err := h.execGoModTidy(outDir)
		panicfIfError(err, "failed to run go mod tidy")
	}

	execFile, err := filepath.Abs(filepath.Join(outDir, cfg.ExecutableFileName()))
	panicfIfError(err, "failed to get absolute path")

	err = h.execGoBuild(append(input.Opt_GoBuildArg, "-o", execFile), packageArgs, outDir)
	panicfIfError(err, "failed to run go build")

	err = h.execBuiltFile(input, execFile)
	panicfIfError(err, "failed to run built file")

	return nil
}

func generateUniqueString(seed int64) string {
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	r := rand.New(rand.NewSource(seed))
	v := []byte{}
	for i := 0; i < 8; i++ {
		v = append(v, alphabet[r.Intn(len(alphabet))])
	}
	return string(v)
}

func (h cliHandler) resolvePackage(packageArg string) (xtracego.ResolveType, []string, xtracego.ResolvedPackageFiles, error) {
	resolveType, packageArgs := xtracego.ParsePackageArgs(packageArg)
	pkg, err := xtracego.ResolvePackage(resolveType, packageArgs)
	return resolveType, packageArgs, pkg, err
}

func (h *cliHandler) transformSourceFiles(
	cfg xtracego.Config,
	pkg xtracego.ResolvedPackageFiles,
	outDir string,
	copyOnlyRegexpStrs []string,
	copyOnlyNotRegexpStr string,
) error {
	srcDir, sourceFiles := pkg.PackageDir, append([]string{}, pkg.SourceFiles...)
	if pkg.GoModFile != "" {
		srcDir, sourceFiles = filepath.Dir(pkg.GoModFile), append(sourceFiles, pkg.GoModFile)
	}

	copyOnlyRegexp := []*regexp.Regexp{}
	for _, s := range copyOnlyRegexpStrs {
		re, err := regexp.Compile(s)
		if err != nil {
			return fmt.Errorf("failed to compile regexp '%s': %w", s, err)
		}
		copyOnlyRegexp = append(copyOnlyRegexp, re)
	}
	copyOnlyNotRegexp, err := regexp.Compile(copyOnlyNotRegexpStr)
	if err != nil {
		return fmt.Errorf("failed to compile regexp '%s': %w", copyOnlyNotRegexpStr, err)
	}
	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)
	for _, srcFile := range sourceFiles {
		eg.Go(func() error {
			isCopyOnly := copyOnlyNotRegexp.MatchString(srcFile) ||
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

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

func (h *cliHandler) saveLibraryFiles(cfg xtracego.Config, outDir string) (dst string, err error) {
	dst = filepath.Join(outDir, cfg.LibraryFileName())
	if cfg.PackageName() != "" {
		dst = filepath.Join(outDir, cfg.PackageName(), cfg.LibraryFileName())
	}
	buf := bytes.NewBuffer(nil)
	if err := xtracego.GetXtraceGo(cfg.UniqueString, buf); err != nil {
		return "", fmt.Errorf("failed to generate library: %w", err)
	}
	if err := xtracego.SaveFile(dst, buf.String()); err != nil {
		return "", fmt.Errorf("failed to save library: %w", err)
	}
	return dst, nil
}

func (h cliHandler) execGoModTidy(outDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
	h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go mod download: %w", err)
	}
	return nil
}

func (h cliHandler) execGoBuild(buildArgs []string, packageArgs []string, outDir string) error {
	args := append(append([]string{"build"}, buildArgs...), packageArgs...)
	cmd := exec.Command("go", args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
	h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go build: %w", err)
	}
	return nil
}

func (h cliHandler) execBuiltFile(input Input_Run, execFile string) error {
	cmd := exec.Command(execFile, input.Arg_Arguments...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	h.logf("[exec] %s", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go build: %w", err)
	}
	return nil
}
