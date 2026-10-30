package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"log"
	"os"
)

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
	fatal            = log.Fatal
)

func fatalIfErr(err error, msg string) {
	if err == nil {
		return
	}
	if msg == "" {
		msg = err.Error()
	}
	fatal(msg)
}

func main() {
	cmdLine := flag.NewFlagSet("handcrafted", flag.ExitOnError)
	wantGenerated := cmdLine.Bool("generated", false, "show generated files instead of handcrafted")
	_ = cmdLine.Parse(os.Args[1:]) //nolint:errcheck // err is always nil in mode flag.ExitOnError
	checker := checkHandcrafted
	if *wantGenerated {
		checker = checkGenerated
	}
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		filename := scanner.Text()
		ok, err := checker(filename)
		fatalIfErr(err, "")
		if ok {
			_, err = stdout.Write([]byte(filename + "\n"))
			fatalIfErr(err, "error writing to stdout")
		}
	}
	fatalIfErr(scanner.Err(), "error reading from stdin")
}

func checkGenerated(filename string) (bool, error) {
	return checkFilename(filename, true)
}

func checkHandcrafted(filename string) (bool, error) {
	return checkFilename(filename, false)
}

func checkFilename(filename string, wantGenerated bool) (bool, error) {
	file, err := os.Open(filename) //nolint:gosec // filename comes from stdin
	if err != nil {
		return false, fmt.Errorf("could not open file %s", filename)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // this error isn't important
	}()

	parsed, err := parser.ParseFile(
		token.NewFileSet(),
		filename,
		file,
		parser.ParseComments|parser.PackageClauseOnly|parser.SkipObjectResolution,
	)
	var parseErrs scanner.ErrorList
	if errors.As(err, &parseErrs) {
		// Files without a valid package clause are neither handcrafted nor generated.
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not read file %s: %w", filename, err)
	}

	return ast.IsGenerated(parsed) == wantGenerated, nil
}
