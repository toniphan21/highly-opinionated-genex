package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/alexflint/go-arg"
	"golang.org/x/tools/go/packages"
	chainergen "nhatp.com/go/chainer-gen"
	composergen "nhatp.com/go/composer-gen"
	genlib "nhatp.com/go/gen-lib"
	"nhatp.com/go/gen-lib/cli"
	stringergen "nhatp.com/go/stringer-gen"
)

type Argument struct {
	WorkingDir string `arg:"-w,--working-dir" help:"Working directory" default:"." placeholder:"WORKING_DIR"`
	DryRun     bool   `arg:"-d,--dry-run" help:"Preview changes without writing to disk"`
}

func (a *Argument) ResolveWorkingDir() string {
	if a.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		return wd
	}

	absPath, err := filepath.Abs(a.WorkingDir)
	if err != nil {
		panic(err)
	}
	return absPath
}

const BinaryVersion = "v0.0.0-example"
const BinaryName = "highly-opinionated-genex"

func main() {
	var args Argument
	arg.MustParse(&args)

	workingDir := args.ResolveWorkingDir()

	if args.DryRun {
		slog.Info(cli.ColorBinary(BinaryName) + " " + cli.ColorVersion(BinaryVersion) + " in DRY mode")
	} else {
		slog.Info(cli.ColorBinary(BinaryName) + " " + cli.ColorVersion(BinaryVersion))
	}
	slog.Info(cli.ColorBinary(BinaryName) + " is working on directory: " + cli.ColorInput(workingDir))

	fileManager := genlib.NewFileManager(workingDir)
	var err error

	chainer := chainergen.New(fileManager)
	if err = doGenerate(workingDir, fileManager, func(pkg *packages.Package) error {
		if cf := makeChainerConfigs(pkg); len(cf) != 0 {
			return chainer.Generate(pkg, cf)
		}
		return nil
	}); err != nil {
		panic(err)
	}

	composer := composergen.New(fileManager)
	if err = doGenerate(workingDir, fileManager, func(pkg *packages.Package) error {
		if cf := makeComposerConfigs(pkg); len(cf) != 0 {
			return composer.Generate(pkg, cf)
		}
		return nil
	}); err != nil {
		panic(err)
	}

	stringer := stringergen.New(fileManager)
	if err = doGenerate(workingDir, fileManager, func(pkg *packages.Package) error {
		if cf := makeStringerConfigs(pkg); len(cf) != 0 {
			return stringer.Generate(pkg, cf)
		}
		return nil
	}); err != nil {
		panic(err)
	}

	if args.DryRun {
		slog.Info(cli.ColorBinary(BinaryName) + " is printing generated file content")
		for _, out := range fileManager.Files() {
			cli.PrintFileWithFunction(out.RelPath, []byte(out.Content()), func(l string) {
				slog.Info(l)
			})
		}
	} else {
		slog.Info(cli.ColorBinary(BinaryName) + " is saving generated file to disk")
		for _, out := range fileManager.Files() {
			if err := os.WriteFile(out.FullPath, []byte(out.Content()), 0644); err != nil {
				panic(err)
			}
		}
	}

	slog.Info(cli.ColorGreen("done"))
}

func makeChainerConfigs(pkg *packages.Package) []chainergen.Config {
	r, err := chainergen.NewRegexMatcher("(?i)^.*Op$")
	if err != nil {
		panic(err)
	}

	result := chainergen.Config{
		PackagePath:       pkg.PkgPath,
		Output:            standardOutput(),
		StructNameMatcher: r,
		ChainedMethods: []chainergen.ChainedMethodConfig{
			{
				NameMatcher:  chainergen.NewStringMatcher("validate", true),
				ErrorHandler: chainergen.NewWrapErrorHandler("validate: %w"),
			},
			{
				NameMatcher:  chainergen.NewStringMatcher("authorize", true),
				ErrorHandler: chainergen.NewWrapErrorHandler("authorize: %w"),
			},
			{
				NameMatcher:  chainergen.NewStringMatcher("process", true),
				ErrorHandler: chainergen.NewWrapErrorHandler("process: %w"),
			},
			{
				NameMatcher:  chainergen.NewStringMatcher("handle", true),
				ErrorHandler: chainergen.NewWrapErrorHandler("handle: %w"),
			},
		},
		MethodName: "execute",
	}
	return []chainergen.Config{result}
}

func makeComposerConfigs(pkg *packages.Package) []composergen.Config {
	var structs []string
	pattern, err := regexp.Compile("(?i)^.*Op$")
	if err != nil {
		panic(err)
	}

	for _, obj := range pkg.TypesInfo.Defs {
		typeName, ok := obj.(*types.TypeName)
		if !ok || typeName.IsAlias() {
			continue
		}
		named, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}
		if _, ok := named.Underlying().(*types.Struct); !ok {
			continue
		}
		if !pattern.MatchString(typeName.Name()) {
			continue
		}

		hasMethod := false
		for i := range named.NumMethods() {
			if named.Method(i).Name() == "execute" {
				hasMethod = true
				break
			}
		}

		if hasMethod {
			structs = append(structs, typeName.Name())
		}
	}

	toExportedAs := func(name string) string {
		name = strings.TrimSuffix(name, "Op")
		if name == "" {
			return name
		}
		r := []rune(name)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}

	var receivers []composergen.Receiver
	for _, name := range structs {
		receivers = append(receivers, composergen.Receiver{
			StructName: name,
			FieldName:  name,
			Method:     "execute",
			ExportedAs: toExportedAs(name),
		})
	}

	if len(receivers) == 0 {
		return nil
	}

	result := composergen.Config{
		PackagePath:           pkg.PkgPath,
		Output:                standardOutput(),
		InterfaceName:         "Service",
		ImplementationName:    "serviceImpl",
		ConstructorName:       "newService",
		DependenciesName:      "serviceDeps",
		DependenciesParamName: "deps",
		Receivers:             receivers,
	}
	return []composergen.Config{result}
}

func makeStringerConfigs(pkg *packages.Package) []stringergen.Config {
	// collect int-family kind
	intTypes := make(map[string]bool)
	for _, obj := range pkg.TypesInfo.Defs {
		typeName, ok := obj.(*types.TypeName)
		if !ok || typeName.IsAlias() {
			continue
		}
		named, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}

		underlying, ok := named.Underlying().(*types.Basic)
		if !ok {
			continue
		}

		switch underlying.Kind() {
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			intTypes[typeName.Name()] = true
		default:
			continue
		}
		intTypes[typeName.Name()] = true
	}

	// filter used in consts
	found := make(map[string]bool)
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				vspec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vspec.Names {
					obj, ok := pkg.TypesInfo.Defs[name].(*types.Const)
					if !ok {
						continue
					}
					named, ok := obj.Type().(*types.Named)
					if !ok {
						continue
					}
					typeName := named.Obj().Name()
					if intTypes[typeName] {
						found[typeName] = true
					}
				}
			}
		}
	}

	var result = stringergen.Config{
		PackagePath: pkg.PkgPath,
		Output:      standardOutput(),
		Binary:      "golang.org/x/tools/cmd/stringer@latest",
		LineComment: true,
	}
	for name := range found {
		result.Types = append(result.Types, name)
	}

	return []stringergen.Config{result}
}

func standardOutput() genlib.Output {
	return genlib.Output{
		SourceFileName: "codegen.go",
		TestFileName:   "codegen_test.go",
	}
}

func doGenerate(dir string, fm genlib.FileManager, fn func(p *packages.Package) error) error {
	pkgs, err := genlib.LoadPackagesWithGenFiles(dir, fm)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		if err = fn(pkg); err != nil {
			return err
		}
	}

	return nil
}
