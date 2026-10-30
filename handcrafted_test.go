package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"handcrafted": main,
	})
}

func TestScript(t *testing.T) {
	t.Parallel()

	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"lines2nul": func(ts *testscript.TestScript, neg bool, args []string) {
				if neg || len(args) == 0 {
					ts.Fatalf("usage: lines2nul file...")
				}
				for _, arg := range args {
					data := ts.ReadFile(arg)
					err := os.WriteFile(ts.MkAbs(arg), []byte(strings.ReplaceAll(data, "\n", "\x00")), 0o600)
					ts.Check(err)
				}
			},
		},
	})
}

func TestScanZero(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "trailing nul",
			input: "a.go\x00b.go\x00",
			want:  []string{"a.go", "b.go"},
		},
		{
			name:  "no trailing nul",
			input: "a.go\x00b.go",
			want:  []string{"a.go", "b.go"},
		},
		{
			name:  "newlines in filenames",
			input: "a\nb.go\x00c.go\x00",
			want:  []string{"a\nb.go", "c.go"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			scanner := bufio.NewScanner(strings.NewReader(tc.input))
			scanner.Split(scanZero)
			var got []string
			for scanner.Scan() {
				got = append(got, scanner.Text())
			}
			err := scanner.Err()
			if err != nil {
				t.Fatalf("scanner.Err(): %v", err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCheckFilenameLongComment(t *testing.T) {
	t.Parallel()

	filename := filepath.Join(t.TempDir(), "example.go")
	source := "// " + strings.Repeat("x", 70*1024) + "\npackage example\n"
	err := os.WriteFile(filename, []byte(source), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	generated, err := checkGenerated(filename)
	if err != nil {
		t.Fatalf("checkGenerated: %v", err)
	}
	if generated {
		t.Error("checkGenerated = true, want false")
	}

	handcrafted, err := checkHandcrafted(filename)
	if err != nil {
		t.Fatalf("checkHandcrafted: %v", err)
	}
	if !handcrafted {
		t.Error("checkHandcrafted = false, want true")
	}
}

func TestCheckFilenameReadError(t *testing.T) {
	t.Parallel()

	filename := t.TempDir()
	_, err := checkGenerated(filename)
	if err == nil {
		t.Fatal("checkGenerated succeeded for a directory")
	}

	_, isPathError := errors.AsType[*os.PathError](err)
	if !isPathError {
		t.Errorf("checkGenerated error %q does not wrap the read error", err)
	}
}
