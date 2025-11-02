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
	"strings"
	"time"

	"github.com/Jumpaku/xtracego"
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

	cfg := xtracego.Config{
		TraceStmt: input.Opt_TraceStmt,
		TraceVar:  input.Opt_TraceVar,
		TraceCall: input.Opt_TraceCall,
		TraceCase: input.Opt_TraceCase,
	}
	cfg.GenPrefix(time.Now().Unix())

	resolveType, packageArgs, _, err := xtracego.ParseArgs(input.Arg_Package)
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

	if err := h.transformSourceFiles(cfg, pkg, outDir); err != nil {
		return fmt.Errorf("failed to clone source files: %w", err)
	}

	return nil
}

func (h cliHandler) Run_Build(input Input_Build) error {
	h.verbose = input.Opt_Verbose
	cfg := xtracego.Config{
		TraceStmt: input.Opt_TraceStmt,
		TraceVar:  input.Opt_TraceVar,
		TraceCall: input.Opt_TraceCall,
		TraceCase: input.Opt_TraceCase,
	}
	cfg.GenPrefix(time.Now().Unix())

	resolveType, packageArgs, _, err := xtracego.ParseArgs(input.Arg_Package)
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

	if err := h.transformSourceFiles(cfg, pkg, outDir); err != nil {
		return fmt.Errorf("failed to clone source files: %w", err)
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
		args = append(args, input.Arg_Package...)
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

	cfg := xtracego.Config{
		TraceStmt: input.Opt_TraceStmt,
		TraceVar:  input.Opt_TraceVar,
		TraceCall: input.Opt_TraceCall,
		TraceCase: input.Opt_TraceCase,
	}
	cfg.GenPrefix(time.Now().Unix())

	resolveType, packageArgs, cliArgs, err := xtracego.ParseArgs(input.Arg_PackageAndArguments)
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

	if err := h.transformSourceFiles(cfg, pkg, outDir); err != nil {
		return fmt.Errorf("failed to clone source files: %w", err)
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
		cmd := exec.Command(execFile, cliArgs...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		h.logf("[exec] %s", cmd.String())
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go build: %w", err)
		}
	}

	return nil
}

func (h cliHandler) transformSourceFiles(cfg xtracego.Config, pkg xtracego.ResolvedPackageFiles, outDir string) error {
	srcDir, sourceFiles := pkg.PackageDir, append([]string{}, pkg.SourceFiles...)
	if pkg.GoModFile != "" {
		srcDir, sourceFiles = filepath.Dir(pkg.GoModFile), append(sourceFiles, pkg.GoModFile)
	}

	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)
	for _, srcFile := range sourceFiles {
		eg.Go(func() error {
			relToFile, err := filepath.Rel(srcDir, srcFile)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			dstFile := filepath.Join(outDir, relToFile)

			err = xtracego.TransformFile(srcFile, dstFile, func(r io.Reader, w io.Writer) (err error) {
				h.logf("[rewrite] %s -> %s", srcFile, dstFile)
				if strings.HasSuffix(srcFile, ".go") {
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
					_, err = io.Copy(w, r)
					if err != nil {
						return fmt.Errorf("failed to copy file: %w", err)
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}
