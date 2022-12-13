// Copyright (c) 2016 - 2019 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package main

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
)

var (
	parsedFiles       map[string]*dst.File
	parsedFileSources map[*dst.File]string
	fset              *token.FileSet
)

type compileFlagSet struct {
	Package string `sqflag:"-p"`
	Output  string `sqflag:"-o"`
}

func (f *compileFlagSet) IsValid() bool {
	return f.Package != "" && f.Output != ""
}

func (f *compileFlagSet) String() string {
	return fmt.Sprintf("-p=%q -o=%q", f.Package, f.Output)
}

func parseCompileCommand(args []string) (commandExecutionFunc, error) {
	if len(args) == 0 {
		return nil, errors.New("unexpected number of command arguments")
	}
	flags := &compileFlagSet{}
	parseFlags(flags, args[1:])
	//fmt.Printf("Flags are %v.\n", flags)
	return makeCompileCommandExecutionFunc(flags, args), nil
}

func makeCompileCommandExecutionFunc(flags *compileFlagSet, args []string) commandExecutionFunc {
	return func() ([]string, error) {
		if !flags.IsValid() {
			// Skip when the required set of flags is not valid.
			log.Printf("nothing to do (%s)\n", flags)
			return nil, nil
		}

		pkgPath := flags.Package
		packageBuildDir := filepath.Dir(flags.Output)

		if pkgPath == "runtime" {
			fmt.Printf("PkgPath %v and packageBuildDir %v.\n", pkgPath, packageBuildDir)
			instrument(args, pkgPath, packageBuildDir)
		}

		return nil, nil
	}
}

func instrumentFuncDeclPre(funcDecl *dst.FuncDecl) {
	if funcDecl.Name.Name == "gopanic" {
		fmt.Printf("Pre: The function name is %v.\n", funcDecl.Name)
		fmt.Printf("Pre: The function type param name is %v and obj is %v.\n", funcDecl.Type.Params.List[0].Names[0].Name, funcDecl.Type.Params.List[0].Names[0].Obj.Kind)
		fmt.Printf("Pre: The receiving argumments are %v\n.", funcDecl.Recv)

		stmt := getDstStatement(funcDecl)

		funcDecl.Body.List = append([]dst.Stmt{stmt}, funcDecl.Body.List...)

	}
}

func getDstStatement(funcDecl *dst.FuncDecl) dst.Stmt {
	return &dst.BlockStmt{
		List: []dst.Stmt{
			&dst.ExprStmt{
				X: &dst.CallExpr{
					Fun: dst.NewIdent("print"),
					Args: []dst.Expr{
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: "Rajeev\n",
						},
					},
					Ellipsis: false,
				},
			},
		},
	}
}

func instrumentPre(cursor *dstutil.Cursor) bool {
	switch node := cursor.Node().(type) {
	case *dst.FuncDecl:
		instrumentFuncDeclPre(node)
		// Note that we don't add the file metadata here in order to avoid to
		// infinite traversal because of adding new AST nodes while visiting it.

		// No need to go deeper than function declarations
		return false
	}
	return true
}

func instrumentFilePost(file *dst.File) {
	fmt.Printf("Post: The file name is %v.\n", file.Name)
}

func instrumentPost(cursor *dstutil.Cursor) bool {
	switch node := cursor.Node().(type) {
	case *dst.File:
		instrumentFilePost(node)
	}
	return true
}

func instrument(args []string, pkgPath, packageBuildDir string) ([]string, error) {
	//

	// Make the list of Go files to instrument out of the argument list and
	// replace their argument list entry by their instrumented copy.
	argIndices := parseCompileCommandArgs(args)
	for src := range argIndices {
		if strings.Contains(src, "panic.go") {
			fmt.Printf("src is %v\n", src)
			AddFile(src)
			root, _ := dst.NewPackage(fset, parsedFiles, nil, nil)
			dstutil.Apply(root, instrumentPre, instrumentPost)
		}
		//if err := i.AddFile(src); err != nil {
		// 	return nil, err
		// }
	}

	// if instrumented, err := i.Instrument(); err != nil {
	// 	return nil, err
	// } else if len(instrumented) > 0 {
	// 	written, err := i.WriteInstrumentedFiles(packageBuildDir, instrumented)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	// Replace original files in the args by the new ones
	// 	updateArgs(args, argIndices, written)
	// }

	// extraFiles, err := i.WriteExtraFiles()
	// if err != nil {
	// 	return nil, err
	// }

	// args = append(args, extraFiles...)
	return args, nil
}

func AddFile(src string) error {
	// Check if the instrumentation should be skipped for this filename
	//fmt.Printf("src is %v\n.", src)

	//fmt.Printf("parsing file %v\n", src)
	fset = token.NewFileSet()
	file, err := decorator.ParseFile(fset, src, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	//fmt.Printf("The decorations in the file are %v.\n", file.Decs)

	// for _, decl := range file.Decls {
	// 	fmt.Printf("Declaration is %v.\n", decl.Decorations())
	// }

	if parsedFiles == nil {
		parsedFiles = make(map[string]*dst.File)
		parsedFileSources = make(map[*dst.File]string)
	}
	parsedFiles[src] = file
	parsedFileSources[file] = src
	return nil
}

func parseCompileCommandArgs(args []string) map[string]int {
	goFiles := make(map[string]int)
	for i, src := range args {
		// Only consider args ending with the Go file extension and assume they
		// are Go files.
		if !strings.HasSuffix(src, ".go") {
			continue
		}
		// Save the position of the source file in the argument list
		// to later change it if it gets instrumented.
		goFiles[src] = i
	}
	return goFiles
}
