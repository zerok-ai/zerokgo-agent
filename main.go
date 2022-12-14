package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

func fix(dir string) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		for fileName, file := range pkg.Files {
			if strings.Contains(fileName, "panic.go") {
				//fmt.Printf("working on file %v\n", fileName)
				ast.Inspect(file, func(n ast.Node) bool {
					fn, ok := n.(*ast.FuncDecl)
					if ok {
						// current node is a function!
						if fn.Name.Name == "gopanic" {
							//fmt.Printf("Func name is %v.\n", fn.Name.Name)
							stmt := []ast.Stmt{
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.Ident{
											Name: "print",
										},
										Lparen: 32,
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: "e",
											},
										},
										Ellipsis: 0,
									},
								},
							}
							fn.Body.List = append(stmt, fn.Body.List...)
						}
					}
					return true
				})

				buf := new(bytes.Buffer)
				err := format.Node(buf, fset, file)
				if err != nil {
					fmt.Printf("error: %v\n", err)
				} else if fileName[len(fileName)-8:] != "_test.go" {
					ioutil.WriteFile(fileName, buf.Bytes(), 0664)
				}
			}
		}
	}
}

func main() {
	fix("/usr/local/Cellar/go/1.19.2/libexec/src/runtime")
}
