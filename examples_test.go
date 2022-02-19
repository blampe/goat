package goat

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

const basePath string = "examples"

func getInOut(t testing.TB, fileName string) (io.Reader, io.Writer) {
	sourceName := filepath.Join(basePath, fileName)
	svgName := filepath.Join(basePath, strings.TrimSuffix(fileName, filepath.Ext(fileName))+".svg")

	in, err := os.Open(sourceName)
	if err != nil {
		t.Error(err)
		return nil, nil
	}

	out, err := os.Create(svgName)
	if err != nil {
		t.Error(err)
		return nil, nil
	}

	return in, out
}

func TestExamplesStableOutput(t *testing.T) {
	c := qt.New(t)

	var previous string
	for i := 0; i < 2; i++ {
		in, err := os.Open(filepath.Join(basePath, "circuits.svg"))
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		ASCIItoSVG(in, &out)
		if i > 0 && previous != out.String() {
			c.Fail()
		}
		previous = out.String()
	}
}

func TestExamples(t *testing.T) {
	fileInfos, err := ioutil.ReadDir(basePath)
	if err != nil {
		t.Error(err)
	}

	for _, fileInfo := range fileInfos {
		in, out := getInOut(t, fileInfo.Name())
		ASCIItoSVG(in, out)
	}
}

func BenchmarkComplicated(b *testing.B) {
	in, out := getInOut(b, "complicated.txt")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ASCIItoSVG(in, out)
	}
}
