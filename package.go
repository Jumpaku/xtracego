package xtracego

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/tools/go/packages"
)

type ResolveType string

const (
	// ResolveTypeUnspecified Cannot resolve package.
	ResolveTypeUnspecified ResolveType = ""

	// ResolveType_CommandLineArguments Source files are specified and go.mod not found.
	// The source files must:
	// - have extension .go,
	// - be in the same directory,
	// - be in the main package,
	// - and contain only one main function.
	ResolveType_CommandLineArguments ResolveType = "command-line-arguments"

	// ResolveType_CommandLineArguments_Module Source files are specified and go.mod found.
	// The source files must:
	// - have extension .go,
	// - be in the same directory,
	// - be in the main package,
	// - and contain only one main function.
	// Multiple source files can be specified with a string of comma-separated source file paths.
	// Dependencies in the same module and external dependencies are resolved via go.mod.
	ResolveType_CommandLineArguments_Module ResolveType = "command-line-arguments(module)"

	// ResolveType_PackageDirectory_Module A directory of a main package is specified and go.mod found.
	// The directory must:
	// - have prefix '/', '.', or '..',
	// - and contain source files which are in the main package and contain only one main function.
	// Dependencies in the same module and external dependencies are resolved via go.mod.
	ResolveType_PackageDirectory_Module ResolveType = "package-directory(module)"
)

type ResolvedPackage struct {
	ResolveType ResolveType
	SourceFiles []string
	PackageDir  string
	GoModFile   string
	Module      string
}

func ResolvePackage(packageArg string) (resolved ResolvedPackage, err error) {
	if packageArg == "" {
		return ResolvedPackage{}, fmt.Errorf("no package specified")
	}
	c := packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedEmbedPatterns | packages.NeedDeps | packages.NeedImports | packages.NeedModule,
	}
	pkgs, err := packages.Load(&c, strings.Split(packageArg, ",")...)
	if err != nil {
		return ResolvedPackage{}, fmt.Errorf("failed to load packages: %w", err)
	}
	var (
		sourceFileSet          = map[string]bool{}
		isCommandLineArguments bool
		mainPackageDir         string
		goModFile              string
		moduleName             string
	)
	for pkg := range packages.Postorder(pkgs) {
		if pkg.Name == "main" {
			if mainPackageDir != "" && mainPackageDir != pkg.Dir {
				return ResolvedPackage{}, fmt.Errorf("multiple main packages found: %q and %q", mainPackageDir, pkg.Dir)
			}
			mainPackageDir = pkg.Dir
		}
		if pkg.PkgPath == "command-line-arguments" {
			isCommandLineArguments = true
		} else {
			if pkg.Module == nil || !pkg.Module.Main {
				continue
			}
		}
		if pkg.Module != nil && pkg.Module.Main {
			goModFile = pkg.Module.GoMod
			moduleName = pkg.Module.Path
		}
		for _, file := range pkg.GoFiles {
			sourceFileSet[file] = true
		}
		for _, pattern := range pkg.EmbedPatterns {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return ResolvedPackage{}, fmt.Errorf("invalid embed pattern: %q", pattern)
			}
			for _, file := range matches {
				sourceFileSet[file] = true
			}
		}
	}
	if mainPackageDir == "" {
		return ResolvedPackage{}, fmt.Errorf("no main package found")
	}
	sourceFiles := lo.Keys(sourceFileSet)
	sort.Strings(sourceFiles)
	if isCommandLineArguments {
		if goModFile == "" {
			return ResolvedPackage{
				ResolveType: ResolveType_CommandLineArguments,
				SourceFiles: sourceFiles,
				PackageDir:  mainPackageDir,
			}, nil
		} else {
			return ResolvedPackage{
				ResolveType: ResolveType_CommandLineArguments_Module,
				SourceFiles: sourceFiles,
				PackageDir:  mainPackageDir,
				GoModFile:   goModFile,
				Module:      moduleName,
			}, nil
		}
	} else {
		return ResolvedPackage{
			ResolveType: ResolveType_PackageDirectory_Module,
			SourceFiles: sourceFiles,
			PackageDir:  mainPackageDir,
			GoModFile:   goModFile,
			Module:      moduleName,
		}, nil
	}
}
