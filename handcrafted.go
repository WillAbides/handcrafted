package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
)

var goListFormat = `{{ range .GoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .TestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .XTestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}`

func main() {
	if filepath.Separator != '/' {
		goListFormat = strings.ReplaceAll(goListFormat, "/", string(filepath.Separator))
	}
	ctx := kong.Parse(&cli)
	ctx.FatalIfErrorf(ctx.Run())
}

var cli struct {
	Filter filterCmd `kong:"cmd,help='filter filenames from stdin removing generated files'"`
	List   listCmd   `kong:"cmd,help='list handcrafted files'"`
}

type filterCmd struct{}

func (f *filterCmd) Run(ctx *kong.Context) error {
	return filterFiles(os.Stdin, ctx.Stdout)
}

type listCmd struct {
	BuildTags  []string `kong:"short=t,help='build tags'"`
	GoListArgs string   `kong:"short=l,help='additional args to pass to go-list'"`
	Packages   []string `kong:"arg,required,help='each package must start with .'"`
	Dir        string   `kong:"type=existingdir,default='.',help='run in this directory'"`
}

func (l *listCmd) Run() error {
	pReader, pWriter := io.Pipe()
	cmd := l.cmd(pWriter)

	filterErrs := make(chan error)
	go func() {
		filterErrs <- filterFiles(pReader, os.Stdout)
	}()
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running go list: %v", err)
	}
	err = pWriter.Close()
	if err != nil {
		panic("PipeWriter.Close always returns nil")
	}
	err = <-filterErrs
	if err != nil {
		return err
	}
	return nil
}

func (l *listCmd) cmd(out io.Writer) *exec.Cmd {
	var listArgs []string
	for _, s := range strings.Split(l.GoListArgs, " ") {
		if s != "" {
			listArgs = append(listArgs, s)
		}
	}
	hasFormat := false
	for _, arg := range listArgs {
		if arg == "-f" {
			hasFormat = true
			break
		}
	}
	if !hasFormat {
		listArgs = append(listArgs, "-f", goListFormat)
	}
	if len(l.BuildTags) > 0 {
		listArgs = append(listArgs, "-tags", strings.Join(l.BuildTags, ","))
	}
	args := append([]string{"list"}, listArgs...)
	args = append(args, l.Packages...)
	cmd := exec.Command("go", args...) //nolint:gosec
	cmd.Stdout = out
	cmd.Dir = l.Dir
	return cmd
}

func filterFiles(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		filename := scanner.Text()
		ok, err := checkFilename(filename)
		if err != nil {
			return err
		}
		if ok {
			_, err = out.Write([]byte(filename + "\n"))
			if err != nil {
				return fmt.Errorf("error writing output")
			}
		}
	}
	return scanner.Err()
}

func checkFilename(filename string) (bool, error) {
	file, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return false, fmt.Errorf("could not open file %s", filename)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // this isn't important
	}()
	res, err := isFileGenerated(file)
	if err != nil {
		return false, fmt.Errorf("error reading file %s", filename)
	}
	return !res, nil
}

// pattern is from https://golang.org/cmd/go/#hdr-Generate_Go_files_by_processing_source
var generatedExp = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func isFileGenerated(rdr io.Reader) (bool, error) {
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		if generatedExp.Match(scanner.Bytes()) {
			return true, nil
		}
	}
	return false, scanner.Err()
}
