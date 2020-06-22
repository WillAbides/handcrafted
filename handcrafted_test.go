package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func setup(t *testing.T, in io.Reader, out io.Writer, args ...string) {
	origStdin := stdin
	stdin = in
	origStdout := stdout
	stdout = out
	origArgs := os.Args
	os.Args = append([]string{"handcrafted"}, args...)
	t.Cleanup(func() {
		stdin = origStdin
		stdout = origStdout
		os.Args = origArgs
	})
}

func withFatal(t *testing.T, fn func(...interface{})) {
	orig := fatal
	fatal = fn
	t.Cleanup(func() {
		fatal = orig
	})
}

func filesList(files ...string) string {
	return strings.Join(files, "\n") + "\n"
}

func Test_main(t *testing.T) {
	t.Run("generated", func(t *testing.T) {
		input := strings.NewReader(filesList("./handcrafted.go", "./handcrafted_test.go", "./generated.go"))
		want := filesList("./generated.go")
		out := new(bytes.Buffer)
		setup(t, input, out, "-generated")
		main()
		requireString(t, want, out.String())
	})

	t.Run("handcrafted", func(t *testing.T) {
		input := strings.NewReader(filesList("./handcrafted.go", "./handcrafted_test.go", "./generated.go"))
		want := filesList("./handcrafted.go", "./handcrafted_test.go")
		out := new(bytes.Buffer)
		setup(t, input, out)
		main()
		requireString(t, want, out.String())
	})

	t.Run("missing file", func(t *testing.T) {
		input := strings.NewReader(filesList("./handcrafted.go", "./missing.go", "./handcrafted_test.go"))
		out := new(bytes.Buffer)
		setup(t, input, out)
		fatalCalls := 0
		withFatal(t, func(errs ...interface{}) {
			fatalCalls++
			requireInt(t, 1, len(errs))
			requireString(t, "could not open file ./missing.go", errs[0].(string))
		})
		main()
		requireInt(t, 1, fatalCalls)
	})
}

func requireInt(t *testing.T, want, got int) {
	t.Helper()
	if want != got {
		t.Fatalf("wanted %d but got %d", want, got)
	}
}

func requireString(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("wanted %q but got %q", want, got)
	}
}
