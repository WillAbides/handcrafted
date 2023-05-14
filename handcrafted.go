package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

// pattern is from https://golang.org/cmd/go/#hdr-Generate_Go_files_by_processing_source
var generatedExp = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

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

	isGenerated := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if generatedExp.Match(scanner.Bytes()) {
			isGenerated = true
			break
		}
	}
	if scanner.Err() != nil {
		return false, fmt.Errorf("error reading file %s", filename)
	}

	return isGenerated == wantGenerated, nil
}
