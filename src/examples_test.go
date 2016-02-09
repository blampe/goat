package goaat

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const basePath string = "../examples/"

func getInOut(t testing.TB, fileName string) (io.Reader, io.Writer) {

	sourceName := basePath + fileName
	svgName := basePath + strings.TrimSuffix(sourceName, filepath.Ext(fileName)) + ".svg"

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
	in, out := getInOut(b, "complicated1.txt")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ASCIItoSVG(in, out)
	}
}
