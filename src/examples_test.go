package goaat

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blampe/goaat/src"
)

func TestExamples(t *testing.T) {
	basePath := "../examples/"

	fileInfos, err := ioutil.ReadDir(basePath)

	if err != nil {
		t.Error(err)
	}

	for _, fileInfo := range fileInfos {
		sourceName := basePath + fileInfo.Name()
		svgName := basePath + strings.TrimSuffix(sourceName, filepath.Ext(sourceName)) + ".svg"

		in, err := os.Open(sourceName)

		if err != nil {
			t.Error(err)
		}

		out, err := os.Create(svgName)

		if err != nil {
			t.Error(err)
		}

		goaat.ASCIItoSVG(in, out)
	}
}
