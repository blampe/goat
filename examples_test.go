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
	regenerate = flag.Bool("regenerate",
		false, "regenerate reference SVG output files")
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
	if regenerate == nil {
		t.Logf("Verifying equality of current SVG with examples/ references.\n")
	}
	var failures int
	for _, name := range filenames {
		in := getIn(name)
		var out io.WriteCloser
		if *regenerate {
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
			if newStr := buff.String(); newStr != golden {
				// Skip complaint if the modification timestamp of the .txt file
				// source is fresher than that of the .svg?
				//   => NO, Any .txt difference might be an editing mistake.

				t.Logf("Content mismatch for %s. Length was %d, expected %d",
					toSVGFilename(name), buff.Len(), len(golden))
				for i:=0; i<min(len(golden), len(newStr)); i++ {
					if newStr[i] != golden[i] {
						t.Logf("Differing runes at offset %d: new='%#v' reference='%#v'\n",
							i, newStr[i], golden[i])
						break
					}
				}
				t.Logf("Generated contents do not match existing %s",
					toSVGFilename(name))
				failures++
			} else {
				if testing.Verbose() {
					t.Logf("Existing and generated contents match %s\n",
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
			"$ go test -run TestExamples -v -args -regenerate")
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

func getIn(txtFilename string) io.ReadCloser {
	in, err := os.Open(txtFilename)
	if err != nil {
		panic(err)
	}
	return in
}

func getOut(txtFilename string) io.WriteCloser {
	out, err := os.Create(toSVGFilename(txtFilename))
	if err != nil {
		panic(err)
	}
	return out
}

func getOutString(txtFilename string) (string, error) {
	b, err := ioutil.ReadFile(toSVGFilename(txtFilename))
	if err != nil {
		return "", err
	}
	// XX  Why are there RETURN characters in contents of the .SVG files?
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return string(b), nil
}

func toSVGFilename(txtFilename string) string {
	return strings.TrimSuffix(txtFilename, filepath.Ext(txtFilename)) + ".svg"
}
