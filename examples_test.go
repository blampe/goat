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

var write = flag.Bool("write", false, "write examples to disk")

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
	c := qt.New(t)

	filenames, err := filepath.Glob(filepath.Join(basePath, "*.txt"))
	c.Assert(err, qt.IsNil)

	var buff *bytes.Buffer

	for _, name := range filenames {
		in := getIn(name)
		if testing.Verbose() {
			t.Logf("\tprocessing %s\n", name)
		}
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

		svgColorLightScheme := "#323232"
		svgColorDarkScheme := "#C8C8C8"
		BuildAndWriteSVG(in, out, svgColorLightScheme, svgColorDarkScheme)

		in.Close()
		out.Close()

		if buff != nil {
			golden, err := getOutString(name)
			if err != nil {
				t.Logf("%s %s:\n\t%s\nConsider:\n\t%s",
					"Option -write not set, and Error reading",
					name,
					err.Error(),
					"$ go test -run TestExamples -v -args -write")
				t.FailNow()
			}
			if buff.String() != golden {
				c.Log(buff.Len(), len(golden))
				c.Fatalf("Content mismatch for %s", name)

			}
			in.Close()
			out.Close()
		}
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
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return string(b), nil
}

func toSVGFilename(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename)) + ".svg"
}
