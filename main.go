package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
)

type params struct {
	Prefix     string `arg:"--prefix,required" help:"The prefix of the generated code eg. Provide"`
	Module     string `arg:"--module,required" help:"The module of the generated code eg. ./service"`
	SetDir     string `arg:"--dir,required" help:"The directory of the generated code eg. ./di"`
	SetFile    string `arg:"--filename" help:"[Optional] The file name of the generated code eg. appset.go"`
	ModulePath string `args:"" help:"[Optional] The module path of the generated code eg. ./service"`
	DiPkg      string `args:"required" help:"The package name of the generated code eg. di"`
}

func main() {
	var args params
	arg.MustParse(&args)

	set := token.NewFileSet()
	if strings.ToUpper(args.Prefix[:1]) != args.Prefix[:1] {
		panic("Prefix must be exported")
	}
	strings.CutSuffix(args.Module, "/")
	if args.ModulePath == "" {
		args.ModulePath = "."
	}
	packs, err := parser.ParseDir(set, args.ModulePath, nil, 0)
	if err != nil {
		panic(err)
	}

	var funcs []*ast.FuncDecl
	for _, pack := range packs {
		if strings.HasSuffix(pack.Name, "_test") {
			continue
		}
		for _, file := range pack.Files {
			for _, decl := range file.Decls {
				if f, ok := decl.(*ast.FuncDecl); ok {
					if strings.HasPrefix(f.Name.Name, args.Prefix) {
						funcs = append(funcs, f)
					}
				}
			}
		}

		fmt.Printf("Found %d Provide functions\n", len(funcs))
		for _, f := range funcs {
			fmt.Printf("%s\n", f.Name.Name)
		}

		fmt.Printf("Found %d packages\n", len(packs))
		for _, pack := range packs {
			fmt.Printf("%s\n", pack.Name)
		}

		// build the wire set
		body := fmt.Sprintf("package %s\n\n", args.DiPkg)
		body += "import (\n"
		for _, pack := range packs {
			body += fmt.Sprintf("\t\"%s/%s\"\n", args.Module, pack.Name)
		}
		body += "\t\"github.com/google/wire\"\n)\n\n"

		body += fmt.Sprintf("var %sSet = wire.NewSet(\n", pack.Name)
		for _, f := range funcs {
			body += fmt.Sprintf("\t%s.%s,\n", pack.Name, f.Name.Name)
		}
		body += ")\n"

		// write to file
		if args.SetFile == "" {
			args.SetFile = pack.Name
		} else {
			args.SetFile = strings.TrimSuffix(args.SetFile, ".go")
		}
		fileName := fmt.Sprintf("%s/%s_set_gen.go", args.SetDir, args.SetFile)
		err := os.WriteFile(fileName, []byte(body), 0644)
		if err != nil {
			return
		}

	}
}
