package main

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		Dir:  ".",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		fmt.Println("Package:", pkg.PkgPath)

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {

				// Structs
				if ts, ok := n.(*ast.TypeSpec); ok {
					if _, ok := ts.Type.(*ast.StructType); ok {
						fmt.Println("  Struct:", ts.Name.Name)
					}
				}

				// Functions + Methods
				if fn, ok := n.(*ast.FuncDecl); ok {
					if fn.Recv != nil {
						fmt.Println("  Method:", fn.Name.Name)
					} else {
						fmt.Println("  Function:", fn.Name.Name)
					}
				}

				return true
			})
		}
	}
}
