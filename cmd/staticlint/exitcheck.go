package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// runExitCheck implements the logic for the exitcheck analyzer.
// It scans the AST for direct calls to os.Exit in the main function
// of the main package and reports them as violations.
//
// The analyzer helps enforce better coding practices by preventing
// direct os.Exit calls that can bypass proper cleanup and make
// code less testable.
func runExitCheck(pass *analysis.Pass) (interface{}, error) {
	// Only check main packages
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Walk through all files in the package
	for _, file := range pass.Files {
		// Walk through all declarations in the file
		ast.Inspect(file, func(node ast.Node) bool {
			// Look for function declarations
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Check if this is the main function
			if funcDecl.Name.Name != "main" {
				return true
			}

			// Check if this function has a receiver (methods can't be main)
			if funcDecl.Recv != nil {
				return true
			}

			// Walk through the function body looking for os.Exit calls
			ast.Inspect(funcDecl, func(node ast.Node) bool {
				callExpr, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if this is a selector expression (package.function)
				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// Check if the selector is "Exit"
				if selExpr.Sel.Name != "Exit" {
					return true
				}

				// Check if the package is "os"
				ident, ok := selExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				if ident.Name == "os" {
					// Report the violation
					pass.Reportf(callExpr.Pos(),
						"direct call to os.Exit in main function is prohibited; "+
							"use proper error handling and graceful shutdown instead")
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
