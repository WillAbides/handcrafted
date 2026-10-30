package main

import (
	"errors"
	"os"
	"path/filepath"
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
	})
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
