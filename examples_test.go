package goat

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

var (
	write = flag.Bool("write", false, "write examples to disk")  // XX rename: more descriptive
	svgColorLightScheme = flag.String("svg-color-light-scheme", "#000000",
		`See help for cmd/goat`)
	svgColorDarkScheme = flag.String("svg-color-dark-scheme", "#FFFFFF",
		`See help for cmd/goat`)
)

// XX  TXT source file suite is limited to a single file -- "circuits.txt"
func TestExamplesStableOutput(t *testing.T) {
	c := qt.New(t)

	var previous string
	for i := 0; i < 3; i++ {
		in, err := os.Open(filepath.Join(basePath, "circuits.txt"))
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		BuildAndWriteSVG(in, &out, "black", "white")
		in.Close()
		if i > 0 && previous != out.String() {
			c.Fail()
		}
		previous = out.String()

	}
}

func TestExamples(t *testing.T) {
	filenames, err := filepath.Glob(filepath.Join(basePath, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}

	var buff *bytes.Buffer
	if write == nil {
		t.Logf("Verifying output of current build against earlier .svg files in examples/.\n")
	}
	var failures int
	for _, name := range filenames {
		in := getIn(name)
		var out io.WriteCloser
		if *write {
			out = getOut(name)
		} else {
			if buff == nil {
				buff = &bytes.Buffer{}
			} else {
				buff.Reset()
			}
			out = struct {
				io.Writer
				io.Closer
			}{
				buff,
				io.NopCloser(nil),
			}
		}

		BuildAndWriteSVG(in, out, *svgColorLightScheme, *svgColorDarkScheme)

		in.Close()
		out.Close()

		if buff != nil {
			golden, err := getOutString(name)
			if err != nil {
				t.Log(err)
			}
			if buff.String() != golden {
				// XX  Skip this if the modification timestamp of the .txt file
				//     source is fresher than the .svg?
				t.Log(buff.Len(), len(golden))
				t.Logf("Content mismatch for %s", toSVGFilename(name))
				failures++
			} else {
				if testing.Verbose() {
					t.Logf("Verified contents of SVG file %s\n",
						toSVGFilename(name))
				}
			}
			in.Close()
			out.Close()
		}
	}
	if failures > 0 {
		t.Logf(`Failed to verify contents of %d .svg files
Consider:
	%s`,
			failures,
			"$ go test -run TestExamples -v -args -write")
		t.FailNow()
	}
}

func BenchmarkComplicated(b *testing.B) {
	in := getIn(filepath.FromSlash("examples/complicated.txt"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildAndWriteSVG(in, io.Discard, "black", "white")
	}
}

const basePath string = "examples"

func getIn(filename string) io.ReadCloser {
	in, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return in
}

func getOut(filename string) io.WriteCloser {
	out, err := os.Create(toSVGFilename(filename))
	if err != nil {
		panic(err)
	}
	return out
}

func getOutString(filename string) (string, error) {
	b, err := ioutil.ReadFile(toSVGFilename(filename))
	if err != nil {
		return "", err
	}
	// XX  Why are there RETURN characters in contents of the .SVG files?
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return string(b), nil
}

func toSVGFilename(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename)) + ".svg"
}
