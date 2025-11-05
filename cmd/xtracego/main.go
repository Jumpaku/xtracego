package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
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

func (h *cliHandler) logf(format string, args ...any) {
	if h.verbose {
		log.Printf(format+"\n", args...)
	}
}

type cliHandler struct {
	verbose bool
}

var _ CLIHandler = &cliHandler{}

func (h *cliHandler) Run(input Input) error {
	log.Println(GetDoc(input.Subcommand))
	return nil
}

func (h *cliHandler) Run_Rewrite(input Input_Rewrite) (err error) {
	h.verbose = input.Opt_Verbose

	resolveType, packageArgs, _, err := xtracego.ParseArgs([]string{input.Arg_Package})
	if err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	outDir := input.Opt_OutputDirectory
	if outDir == "" {
		return fmt.Errorf("-output-directory is required")
	}

	pkg, err := xtracego.ResolvePackage(resolveType, packageArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve package: %w", err)
	}
	if pkg.GoModFile == "" {
		return fmt.Errorf("go.mod is required")
	}

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   resolveType,
		ModulePath:    pkg.Module,
	}
	cfg.GenUniqueString(time.Now().Unix())

	if err := h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot); err != nil {
		return fmt.Errorf("failed to clone source files: %w", err)
	}

	if err := h.saveLibraryFiles(cfg, outDir); err != nil {
		return fmt.Errorf("failed to save xtracego library: %w", err)
	}

	return nil
}

func (h cliHandler) Run_Build(input Input_Build) error {
	h.verbose = input.Opt_Verbose

	resolveType, packageArgs, _, err := xtracego.ParseArgs([]string{input.Arg_Package})
	if err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	outDir := input.Opt_BuildDirectory
	if outDir == "" {
		return fmt.Errorf("-build-directory is required")
	}

	pkg, err := xtracego.ResolvePackage(resolveType, packageArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve package: %w", err)
	}
	if pkg.GoModFile == "" {
		return fmt.Errorf("go.mod is required")
	}

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   resolveType,
		ModulePath:    pkg.Module,
	}
	cfg.GenUniqueString(time.Now().Unix())

	if err := h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot); err != nil {
		return fmt.Errorf("failed to clone source files: %w", err)
	}

	if err := h.saveLibraryFiles(cfg, outDir); err != nil {
		return fmt.Errorf("failed to save xtracego library: %w", err)
	}

	if pkg.GoModFile != "" {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
		h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go mod download: %w", err)
		}
	}
	{
		args := []string{"build"}
		args = append(args, input.Opt_GoBuildArg...)
		args = append(args, input.Arg_Package)
		cmd := exec.Command("go", args...)
		cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
		h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go build: %w", err)
		}
	}

	return nil
}

func (h cliHandler) Run_Run(input Input_Run) error {
	h.verbose = input.Opt_Verbose
	resolveType, packageArgs, _, err := xtracego.ParseArgs([]string{input.Arg_Package})
	if err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	outDir, err := os.MkdirTemp("", "xtracego_*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(outDir)

	pkg, err := xtracego.ResolvePackage(resolveType, packageArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve package: %w", err)
	}
	if pkg.GoModFile == "" {
		return fmt.Errorf("go.mod is required")
	}

	cfg := xtracego.Config{
		TraceStmt:     input.Opt_TraceStmt,
		TraceVar:      input.Opt_TraceVar,
		TraceCall:     input.Opt_TraceCall,
		ShowTimestamp: input.Opt_Timestamp,
		ShowGoroutine: input.Opt_Goroutine,
		ResolveType:   resolveType,
		ModulePath:    pkg.Module,
	}
	cfg.GenUniqueString(time.Now().Unix())

	if err := h.transformSourceFiles(cfg, pkg, outDir, input.Opt_CopyOnly, input.Opt_CopyOnlyNot); err != nil {
		return fmt.Errorf("failed to clone source files: %w", err)
	}

	if err := h.saveLibraryFiles(cfg, outDir); err != nil {
		return fmt.Errorf("failed to save xtracego library: %w", err)
	}

	if pkg.GoModFile != "" {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
		h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go mod download: %w", err)
		}
	}

	execFile, err := filepath.Abs(filepath.Join(outDir, "main"))
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	{
		args := []string{"build"}
		args = append(args, input.Opt_GoBuildArg...)
		args = append(args, "-o", execFile)
		args = append(args, packageArgs...)
		cmd := exec.Command("go", args...)
		cmd.Dir, cmd.Stdout, cmd.Stderr, cmd.Stdin = outDir, os.Stdout, os.Stderr, os.Stdin
		h.logf("[exec] %s [%s]", cmd.String(), cmd.Dir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go build: %w", err)
		}
	}
	{
		cmd := exec.Command(execFile, input.Arg_Arguments...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		h.logf("[exec] %s", cmd.String())
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go build: %w", err)
		}
	}

	return nil
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

func (h *cliHandler) saveLibraryFiles(cfg xtracego.Config, outDir string) (err error) {
	dst := filepath.Join(outDir, cfg.FileName())
	if cfg.PackageName() != "" {
		dst = filepath.Join(outDir, cfg.PackageName(), cfg.FileName())
	}
	buf := bytes.NewBuffer(nil)
	if err := xtracego.GetXtraceGo(cfg.UniqueString, buf); err != nil {
		return fmt.Errorf("failed to generate library: %w", err)
	}
	if err := xtracego.SaveFile(dst, buf.String()); err != nil {
		return fmt.Errorf("failed to save library: %w", err)
	}
	return nil
}
