package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Jumpaku/goxe"
	"golang.org/x/sync/errgroup"
)

//go:generate cyamli generate golang -schema-path=cli.yaml -out-path=cli.gen.go
func main() {
	if err := Run(cliHandler{}, os.Args); err != nil {
		log.Panicf("error: %+v\n", err)
	}
}

type cliHandler struct{}

var _ CLIHandler = cliHandler{}

func (h cliHandler) Run(input Input) error {
	log.Println(GetDoc(input.Subcommand))
	return nil
}

func (h cliHandler) Run_Rewrite(input Input_Rewrite) (err error) {
	ctx := context.Background()

	resolveType, packageArgs, _, err := goxe.ParseArgs(input.Arg_Package)
	if err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	pkg, err := goxe.ResolvePackage(resolveType, packageArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve package: %w", err)
	}

	outDir := input.Opt_OutputDirectory
	if outDir == "" {
		outDir, err = os.MkdirTemp("", "goxe_*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(outDir)
	}

	if err := cloneSourceFiles(ctx, pkg, outDir); err != nil {
		return fmt.Errorf("failed to clone source files: %w", err)
	}

	return nil
}

func (h cliHandler) Run_Build(input Input_Build) error {
	//TODO implement me
	panic("implement me")
}

func (h cliHandler) Run_Run(input Input_Run) error {
	//TODO implement me
	panic("implement me")
}

func cloneSourceFiles(ctx context.Context, pkg goxe.ResolvedPackageFiles, outDir string) error {
	srcDir, sourceFiles := pkg.PackageDir, append([]string{}, pkg.SourceFiles...)
	if pkg.GoModFile != "" {
		srcDir, sourceFiles = filepath.Dir(pkg.GoModFile), append(sourceFiles, pkg.GoModFile)
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, srcFile := range pkg.SourceFiles {
		eg.Go(func() error {
			relToFile, err := filepath.Rel(srcDir, srcFile)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			dstFile := filepath.Join(outDir, relToFile)

			err = goxe.TransformFile(srcFile, dstFile, func(r io.Reader, w io.Writer) (err error) {
				if _, err = io.Copy(w, r); err != nil {
					return fmt.Errorf("failed to copy file: %w", err)
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
