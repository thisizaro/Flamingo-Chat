package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	root := "."

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".go" {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		fmt.Println("File:", path)

		for _, decl := range node.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				fmt.Println("  -", fn.Name.Name)
			}
		}

		return nil
	})
}	
