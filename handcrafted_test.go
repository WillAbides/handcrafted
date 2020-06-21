package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return wd
}

func Test_filterFiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		wd := pwd(t)
		want := []string{
			filepath.Join(wd, "handcrafted.go"),
			filepath.Join(wd, "handcrafted_test.go"),
		}
		inFiles := append(want, filepath.Join(wd, "generated.go"))
		input := strings.NewReader(strings.Join(inFiles, "\n") + "\n")
		var out bytes.Buffer
		err := filterFiles(input, &out)
		require.NoError(t, err)
		got := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
		sort.Strings(got)
		require.Equal(t, want, got)
	})
}

func Test_checkFilename(t *testing.T) {
	t.Run("handcrafted", func(t *testing.T) {
		filename := filepath.Join(pwd(t), "handcrafted_test.go")
		got, err := checkFilename(filename)
		require.NoError(t, err)
		require.True(t, got)
	})

	t.Run("generated", func(t *testing.T) {
		filename := filepath.Join(pwd(t), "generated.go")
		got, err := checkFilename(filename)
		require.NoError(t, err)
		require.False(t, got)
	})

	t.Run("non-existant file", func(t *testing.T) {
		filename := filepath.Join(pwd(t), "fake.go")
		got, err := checkFilename(filename)
		require.Error(t, err)
		require.False(t, got)
	})
}

func Test_isFileGenerated(t *testing.T) {
	t.Run("generated", func(t *testing.T) {
		content, err := ioutil.ReadFile("generated.go")
		require.NoError(t, err)
		got, err := isFileGenerated(bytes.NewReader(content))
		require.NoError(t, err)
		require.True(t, got)
	})

	t.Run("handcrafted", func(t *testing.T) {
		content := `
package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
`
		got, err := isFileGenerated(strings.NewReader(content))
		require.NoError(t, err)
		require.False(t, got)
	})

	t.Run("error rdr", func(t *testing.T) {
		in := errReader{
			reader: strings.NewReader("package main"),
			err:    assert.AnError,
		}
		got, err := isFileGenerated(&in)
		require.EqualError(t, err, assert.AnError.Error())
		require.False(t, got)
	})
}

type errReader struct {
	reader io.Reader
	err    error
}

func (e *errReader) Read(p []byte) (n int, err error) {
	got, err := e.reader.Read(p)
	if err == io.EOF {
		err = e.err
	}
	return got, err
}
