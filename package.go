package xtracego

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

type ResolveType string

const (
	ResolveTypeUnspecified          ResolveType = ""
	ResolveTypeCommandLineArguments ResolveType = "command-line-arguments"
	ResolveTypePackageDirectory     ResolveType = "package-directory"
)

type ResolvedPackageFiles struct {
	ResolveType ResolveType
	SourceFiles []string
	GoModFile   string
	PackageDir  string
	Module      string
}

func (pkg ResolvedPackageFiles) IsModule() bool {
	return pkg.GoModFile != ""
}

func ParsePackageArgs(packageArg string) (resolveType ResolveType, packageArgs []string) {
	if info, err := os.Stat(packageArg); err == nil && info.IsDir() {
		return ResolveTypePackageDirectory, []string{packageArg}
	}
	return ResolveTypeCommandLineArguments, strings.Split(packageArg, ",")

}

func ResolvePackage(resolveType ResolveType, packageArgs []string) (resolved ResolvedPackageFiles, err error) {
	switch resolveType {
	default:
		return ResolvedPackageFiles{}, fmt.Errorf("invalid resolve type: %q", resolveType)
	case ResolveTypeCommandLineArguments:
		{
			// all files must be .go files in the same directory.
			if err := validateCommandLineArguments(packageArgs); err != nil {
				return ResolvedPackageFiles{}, err
			}
			file, err := filepath.Abs(packageArgs[0])
			if err != nil {
				return ResolvedPackageFiles{}, fmt.Errorf("failed to resolve %q: %w", packageArgs[0], err)
			}
			goModPath, moduleName, goModPathFound, err := findGoMod(filepath.Dir(file))
			if err != nil {
				return ResolvedPackageFiles{}, fmt.Errorf("failed to find go.mod: %w", err)
			}
			sourceFiles, err := resolveSourceFiles(packageArgs)
			if err != nil {
				return ResolvedPackageFiles{}, fmt.Errorf("failed to resolve source files: %w", err)
			}
			resolved := ResolvedPackageFiles{ResolveType: resolveType, SourceFiles: sourceFiles, PackageDir: filepath.Dir(file)}
			if goModPathFound {
				resolved.GoModFile = goModPath
			}
			if moduleName != "" {
				resolved.Module = moduleName
			}
			return resolved, nil
		}
	case ResolveTypePackageDirectory:
		{
			if len(packageArgs) == 0 {
				return ResolvedPackageFiles{}, fmt.Errorf("no package specified")
			}
			packageDir := packageArgs[0]
			goModPath, moduleName, found, err := findGoMod(packageDir)
			if err != nil {
				return ResolvedPackageFiles{}, fmt.Errorf("failed to find go.mod: %w", err)
			}
			sourceFiles, err := resolveSourceFiles(packageArgs)
			if err != nil {
				return ResolvedPackageFiles{}, fmt.Errorf("failed to resolve source files: %w", err)
			}
			resolved := ResolvedPackageFiles{ResolveType: resolveType, SourceFiles: sourceFiles, PackageDir: packageDir}
			if found {
				resolved.GoModFile = goModPath
			}
			if moduleName != "" {
				resolved.Module = moduleName
			}
			return resolved, nil
		}
	}
}

func validateCommandLineArguments(packageArgs []string) error {
	if len(packageArgs) == 0 {
		return fmt.Errorf("no package specified")
	}
	dirs := map[string]bool{}
	for _, file := range packageArgs {
		absFile, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to resolve %q: %w", file, err)
		}
		s, err := os.Stat(absFile)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %w", file, err)
		}
		if s.IsDir() {
			return fmt.Errorf("all files must be .go files, got directory %q", file)
		}

		dirs[filepath.Dir(absFile)] = true
	}
	if len(dirs) != 1 {
		return fmt.Errorf("all files must be in the same directory")
	}
	return nil
}

func findGoMod(dir string) (goModPath, moduleName string, found bool, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to resolve %q: %w", dir, err)
	}

	for {
		goModPath = filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err != nil {
			if !os.IsNotExist(err) {
				return "", "", false, fmt.Errorf("failed to stat %q: %w", goModPath, err)
			}
		} else {
			data, err := os.ReadFile(goModPath)
			if err != nil {
				return "", "", false, fmt.Errorf("failed to read %q: %w", goModPath, err)
			}
			modFile, err := modfile.Parse(goModPath, data, nil)
			if err != nil {
				return "", "", false, fmt.Errorf("failed to parse %q: %w", goModPath, err)
			}
			if modFile.Module == nil || modFile.Module.Mod.Path == "" {
				return "", "", false, fmt.Errorf("invalid module in %q", goModPath)
			}
			return goModPath, modFile.Module.Mod.Path, true, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", false, nil
		}
		dir = parent
	}
}

func resolveSourceFiles(packageArgs []string) (sourceFiles []string, err error) {
	c := packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedEmbedPatterns | packages.NeedDeps | packages.NeedImports | packages.NeedModule,
	}
	pkgs, err := packages.Load(&c, packageArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	sourceFileSet := map[string]bool{}
	mainPkgs := map[string]bool{}
	for pkg := range packages.Postorder(pkgs) {
		if pkg.Name == "main" {
			mainPkgs[pkg.PkgPath] = true
		}
		if pkg.PkgPath != "command-line-arguments" {
			if pkg.Module == nil || !pkg.Module.Main {
				continue
			}
		}

		for _, file := range pkg.GoFiles {
			sourceFileSet[file] = true
		}
		for _, pattern := range pkg.EmbedPatterns {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid embed pattern: %q", pattern)
			}
			for _, file := range matches {
				sourceFileSet[file] = true
			}
		}
	}
	sourceFiles = lo.Keys(sourceFileSet)
	sort.Strings(sourceFiles)

	if len(mainPkgs) != 1 {
		return nil, fmt.Errorf("main packages must be exactly one, got %d", len(mainPkgs))
	}

	return sourceFiles, nil
}
