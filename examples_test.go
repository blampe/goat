package goat

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	write = flag.Bool("write",
		false, "write reference SVG output files")
	svgColorLightScheme = flag.String("svg-color-light-scheme", "#000000",
		`See help for cmd/goat`)
	svgColorDarkScheme = flag.String("svg-color-dark-scheme", "#FFFFFF",
		`See help for cmd/goat`)
)

// XX  TXT source file suite is limited to a single file -- "circuits.txt"
func TestExampleStableOutput(t *testing.T) {
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
			t.FailNow()
		}
		previous = out.String()

	}
}

func TestExamples(t *testing.T) {
	filenames, err := filepath.Glob(filepath.Join(basePath, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}

	if *write {
		writeExamples(t, filenames)
	} else {
		t.Logf("Verifying equality of current SVG with examples/ references.\n")
		verifyExamples(t, filenames)
	}
}


func writeExamples(t *testing.T, filenames []string) {
	for _, name := range filenames {
		in := getIn(name)
		out := getOut(name)
		BuildAndWriteSVG(in, out, *svgColorLightScheme, *svgColorDarkScheme)
		in.Close()
		out.Close()
	}
}

func verifyExamples(t *testing.T, filenames []string) {
	var failures []string
	for _, name := range filenames {
		in := getIn(name)
		buff := &bytes.Buffer{}
		BuildAndWriteSVG(in, buff, *svgColorLightScheme, *svgColorDarkScheme)
		in.Close()
		if nil != compareSVG(t, buff, name) {
			failures = append(failures, name)
		}

	}
	if len(failures) > 0 {
		t.Logf(`Failed to verify contents of %d .svg files
Failing files:`,
			len(failures))
		for _, name := range failures {
			svgFile := toSVGFilename(name)
			fmt.Printf("\t\t%s\n", svgFile)
		}
		t.FailNow()
	}
}

func compareSVG(t *testing.T, buff *bytes.Buffer, fileName string) error {
	golden, err := getOutString(fileName)
	if err != nil {
		t.Log(err)
	}
	if newStr := buff.String(); newStr != golden {
		// Skip complaint if the modification timestamp of the .txt file
		// source is fresher than that of the .svg?
		//   => NO, Any .txt difference might be an editing mistake.

		t.Logf("Content mismatch for %s. Length was %d, expected %d",
			toSVGFilename(fileName), buff.Len(), len(golden))
		for i:=0; i<min(len(golden), len(newStr)); i++ {
			if newStr[i] != golden[i] {
				t.Logf("Differing runes at offset %d: new='%#v' reference='%#v'\n",
					i, newStr[i], golden[i])
				break
			}
		}
		t.Logf("Generated contents do not match existing %s",
			toSVGFilename(fileName))
		return errors.New("Generated contents do not match existing")
	} else {
		if testing.Verbose() {
			t.Logf("Existing and generated contents match %s\n",
				toSVGFilename(fileName))
		}
	}
	return nil
}

func BenchmarkComplicated(b *testing.B) {
	in := getIn(filepath.FromSlash("examples/complicated.txt"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildAndWriteSVG(in, io.Discard, "black", "white")
	}
	in.Close()
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
		// XX  Simply panic rather than return an error?
		return "", err
	}
	// XX  Why are there RETURN characters in contents of the .SVG files?
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return string(b), nil
}

func toSVGFilename(txtFilename string) string {
	return strings.TrimSuffix(txtFilename, filepath.Ext(txtFilename)) + ".svg"
}
