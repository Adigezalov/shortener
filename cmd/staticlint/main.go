// Package staticlint provides a custom static analysis tool (multichecker) that combines
// multiple analyzers for comprehensive Go code analysis.
//
// # Multichecker Overview
//
// This multichecker combines several categories of static analyzers:
//   - Standard analyzers from golang.org/x/tools/go/analysis/passes
//   - All SA class analyzers from staticcheck.io
//   - Additional analyzers from other staticcheck.io classes
//   - Public third-party analyzers
//   - Custom analyzers for project-specific rules
//
// # Usage
//
// To run the static analysis on your project:
//
//	go run ./cmd/staticlint ./...
//
// Or build and run:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// # Included Analyzers
//
// ## Standard Analyzers (golang.org/x/tools/go/analysis/passes)
//
// - asmdecl: reports mismatches between assembly files and Go declarations
// - assign: detects useless assignments
// - atomic: checks for common mistakes using the sync/atomic package
// - bools: detects common mistakes involving boolean operators
// - buildtag: checks that +build tags are well-formed and correctly located
// - cgocall: detects some violations of the cgo pointer passing rules
// - composite: checks for unkeyed composite literals
// - copylock: checks for locks erroneously passed by value
// - errorsas: checks that the second argument to errors.As is a pointer to a type implementing error
// - httpresponse: checks for mistakes using HTTP responses
// - loopclosure: checks for loops that capture loop variables
// - lostcancel: checks for failure to call a context cancellation function
// - nilfunc: checks for useless comparisons between functions and nil
// - printf: checks consistency of Printf format strings and arguments
// - shadow: checks for possible unintended shadowing of variables
// - shift: checks for shifts that equal or exceed the width of the integer
// - simplifycompositelit: suggests using shorter variable names for composite literals
// - simplifyrange: suggests replacing for loops with range loops where possible
// - simplifyslice: suggests using shorter slice syntax
// - sortslice: checks for incorrect usage of sort.Slice
// - stdmethods: checks for misspellings of well-known method names
// - stringintconv: flags type conversions from integers to strings
// - structtag: checks that struct field tags conform to reflect.StructTag.Get
// - tests: checks for common mistaken usages of tests and benchmarks
// - unmarshal: reports passing non-pointer or non-interface values to unmarshal
// - unreachable: checks for unreachable code
// - unsafeptr: checks for invalid conversions of uintptr to unsafe.Pointer
// - unusedresult: checks for unused results of calls to some functions
//
// ## Staticcheck SA Analyzers
//
// All SA (Static Analysis) class analyzers from staticcheck.io that detect
// various code quality issues, bugs, and potential problems.
//
// ## Additional Staticcheck Analyzers
//
// - S1000-S1999: Code simplification suggestions
// - ST1000-ST1999: Style guide violations
// - QF1000-QF1999: Quick fixes for common issues
//
// ## Third-party Analyzers
//
// - errcheck: checks for unchecked errors
//
// ## Custom Analyzers
//
// - exitcheck: prevents direct os.Exit calls in main function of main package
//
// # Custom Analyzer: exitcheck
//
// The exitcheck analyzer enforces a coding standard that prohibits direct calls
// to os.Exit() in the main function of the main package. This promotes:
//   - Better testability by allowing graceful shutdown
//   - Proper resource cleanup through deferred functions
//   - More predictable application lifecycle management
//
// Instead of os.Exit(), applications should use proper error handling and
// graceful shutdown mechanisms.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/kisielk/errcheck/errcheck"
)

// exitcheckAnalyzer is a custom analyzer that prevents direct os.Exit calls
// in the main function of the main package.
//
// This analyzer promotes better coding practices by encouraging:
//   - Proper error handling and propagation
//   - Graceful shutdown mechanisms
//   - Better testability of main functions
//   - Proper resource cleanup through deferred functions
var exitcheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for direct os.Exit calls in main function of main package",
	Run:  runExitCheck,
}

// main is the entry point for the staticlint multichecker.
// It combines multiple static analysis tools into a single command.
func main() {
	var analyzers []*analysis.Analyzer

	// Add standard analyzers from golang.org/x/tools/go/analysis/passes
	analyzers = append(analyzers,
		asmdecl.Analyzer,       // Check assembly declarations
		assign.Analyzer,        // Check for useless assignments
		atomic.Analyzer,        // Check atomic package usage
		bools.Analyzer,         // Check boolean operators
		buildtag.Analyzer,      // Check build tags
		cgocall.Analyzer,       // Check cgo calls
		composite.Analyzer,     // Check composite literals
		copylock.Analyzer,      // Check for copied locks
		errorsas.Analyzer,      // Check errors.As usage
		httpresponse.Analyzer,  // Check HTTP response usage
		loopclosure.Analyzer,   // Check loop closures
		lostcancel.Analyzer,    // Check for lost context cancellation
		nilfunc.Analyzer,       // Check nil function comparisons
		printf.Analyzer,        // Check printf format strings
		shadow.Analyzer,        // Check for variable shadowing
		shift.Analyzer,         // Check bit shifts
		sortslice.Analyzer,     // Check sort.Slice usage
		stdmethods.Analyzer,    // Check standard method signatures
		stringintconv.Analyzer, // Check string/int conversions
		structtag.Analyzer,     // Check struct tags
		tests.Analyzer,         // Check test functions
		unmarshal.Analyzer,     // Check unmarshal usage
		unreachable.Analyzer,   // Check for unreachable code
		unsafeptr.Analyzer,     // Check unsafe pointer usage
		unusedresult.Analyzer,  // Check for unused results
	)

	// Add all SA class analyzers from staticcheck
	for _, analyzer := range staticcheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// Add analyzers from other staticcheck classes
	// Simple (S) - code simplification
	for _, analyzer := range simple.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// Style (ST) - style guide
	for _, analyzer := range stylecheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// Quick Fix (QF) - quick fixes
	for _, analyzer := range quickfix.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// Add public third-party analyzers
	analyzers = append(analyzers,
		errcheck.Analyzer, // Check for unchecked errors
	)

	// Add our custom analyzer
	analyzers = append(analyzers, exitcheckAnalyzer)

	// Run the multichecker with all analyzers
	multichecker.Main(analyzers...)
}
