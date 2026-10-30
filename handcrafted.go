package main

import (
	"bufio"
	"cmp"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"
	"os"
)

func fatalIfErr(err error, msg string) {
	if err == nil {
		return
	}
	log.Fatal(cmp.Or(msg, err.Error()))
}

func main() {
	cmdLine := flag.NewFlagSet("handcrafted", flag.ExitOnError)
	wantGenerated := cmdLine.Bool("generated", false, "show generated files instead of handcrafted")

	//nolint:errcheck // with flag.ExitOnError, Parse always returns nil error
	_ = cmdLine.Parse(os.Args[1:])

	checker := checkHandcrafted
	if *wantGenerated {
		checker = checkGenerated
	}
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		filename := stdin.Text()
		ok, err := checker(filename)
		fatalIfErr(err, "")
		if ok {
			_, err = os.Stdout.WriteString(filename + "\n")
			fatalIfErr(err, "error writing to stdout")
		}
	}
	fatalIfErr(stdin.Err(), "error reading from stdin")
}

func checkGenerated(filename string) (bool, error) {
	return checkFilename(filename, true)
}

func checkHandcrafted(filename string) (bool, error) {
	return checkFilename(filename, false)
}

func checkFilename(filename string, wantGenerated bool) (bool, error) {
	mode := parser.ParseComments | parser.PackageClauseOnly | parser.SkipObjectResolution
	parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, mode)
	_, isParseError := errors.AsType[scanner.ErrorList](err)
	if isParseError {
		// Files without a valid package clause are neither handcrafted nor generated.
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not read file %s: %w", filename, err)
	}

	return ast.IsGenerated(parsed) == wantGenerated, nil
}
