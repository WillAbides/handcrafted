package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
)

var defaultFormat = `{{ range .GoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .TestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .XTestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}`

func init() {
	defaultFormat = strings.ReplaceAll(defaultFormat, "/", string(filepath.Separator))
}

var cli struct {
	BuildTags  []string `kong:"short=t,help='build tags'"`
	GoListArgs string   `kong:"short=l,help='additional args to pass to go-list'"`
	Packages   []string `kong:"arg,required,help='each package must start with .'"`
	Dir        string   `kong:"type=existingdir,default='.',help='run in this directory'"`
}

func main() {
	kong.Parse(&cli).FatalIfErrorf(run())
}

func run() error {
	allFiles, err := goList()
	if err != nil {
		return err
	}
	for _, filename := range allFiles {
		ok, err := checkFilename(filename)
		if err != nil {
			return err
		}
		if ok {
			fmt.Println(filename)
		}
	}
	return nil
}

func goList() ([]string, error) {
	var listArgs []string
	for _, s := range strings.Split(cli.GoListArgs, " ") {
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
		listArgs = append(listArgs, "-f", defaultFormat)
	}
	args := append([]string{"list"}, listArgs...)
	args = append(args, cli.Packages...)
	cmd := exec.Command("go", args...) //nolint:gosec
	cmd.Dir = cli.Dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running go list: %v", err)
	}
	out = bytes.TrimSpace(out)
	return strings.Split(string(out), "\n"), nil
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
