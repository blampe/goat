package goat

import (
	"bytes"
	"errors"
	"flag"
	//"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

const (
	examplesDir = "examples"
)

var (
	write = flag.Bool("write",
		false, "write reference SVG output files")
	svgColorLightScheme = flag.String("svg-color-light-scheme", "#000000",
		`See help for cmd/goat`)
	svgColorDarkScheme = flag.String("svg-color-dark-scheme", "#FFFFFF",
		`See help for cmd/goat`)

	// Begin the directory name with '_' to hide from git.
	svgDeltaDir = flag.String("svg-delta-dir", "_examples_new",
		`Directory to be filled with a delta-image file for each
newly-generated SVG that does not match those in ` + examplesDir)
)

func TestExamples(t *testing.T) {
	// XX  This sweeps up ~every~ *.txt file in examples/
	txtPaths, err := filepath.Glob(filepath.Join(examplesDir, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}

	baseNames := make([]string, len(txtPaths))
	for i := range txtPaths {
		baseName, found := strings.CutPrefix(txtPaths[i], examplesDir+"/")
		if !found {
			panic("Could not cut prefix from pathname.")
		}
		baseNames[i] = baseName
	}

	if *write {
		writeExamples(examplesDir, examplesDir, baseNames, *svgColorLightScheme, *svgColorDarkScheme)
	} else {
		t.Logf("Verifying equality of current SVG with examples/ references.\n")
		verifyExamples(t, examplesDir, baseNames)
	}
}


func writeExamples(inDir, outDir string, baseNames []string, lightColor, darkColor string) {
	for _, name := range baseNames {
		in := getIn(inDir + "/" + name)
		out := getOut(outDir + "/" + name)
		BuildAndWriteSVG(in, out, lightColor, darkColor)
		in.Close()
		out.Close()
	}
}

func verifyExamples(t *testing.T, examplesDir string, baseNames []string) {
	var failures []string
	for _, name := range baseNames {
		in := getIn(examplesDir + "/" + name)
		buff := &bytes.Buffer{}
		BuildAndWriteSVG(in, buff, *svgColorLightScheme, *svgColorDarkScheme)
		in.Close()
		if nil != compareSVG(t, buff, examplesDir, name) {
			failures = append(failures, name)
		}

	}
	if len(failures) > 0 {
		t.Logf(`Failed to verify contents of %d .svg files`,
			len(failures))
		err := os.Mkdir(*svgDeltaDir, 0770)
		if err != nil {
			t.Fatalf(`
    Aborting: "%v"`, err)
		}
		writeExamples(examplesDir, *svgDeltaDir, failures, "#000088", "#88CCFF")
		writeDeltaHTML(t, "../" + examplesDir, *svgDeltaDir, failures)
		t.FailNow()
	}
}

func compareSVG(t *testing.T, buff *bytes.Buffer, examplesDir string, baseName string) error {
	fileName := examplesDir + "/" + baseName
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

// See https://developer.mozilla.org/en-US/docs/Web/CSS/blend-mode#example_using_difference
func writeDeltaHTML(t *testing.T, examplesDir, deltaDir string, baseNames []string) {
	t.Logf("Writing new SVG and HTML delta files into %s/", deltaDir)

	tmpl := template.Must(template.New("_ignored_").Parse(`
<style type="text/css">
.blended-images {
    height: 100%; /* XX  How to make equal to pixel bounds of the SVGs? */
    background-size: contain, contain;
    background-repeat: no-repeat;
    background-blend-mode: difference;
    background-image: url('{{.ExamplesDir}}/{{.SvgBaseName}}'), url('{{.DeltaDir}}/{{.SvgBaseName}}');
 }
</style>

<div style="background-color: grey;">
    <div class="blended-images"></div>
</div>
`))
	for _, name := range baseNames {
		htmlOutName := stripSuffix(name) + ".html"
		t.Logf("\t%s", htmlOutName)
		htmlOutFile, err := os.Create(deltaDir + "/" + htmlOutName)
		err = tmpl.Execute(htmlOutFile, map[string]string{
			"ExamplesDir": examplesDir,
			"DeltaDir": ".",
			"SvgBaseName": toSVGFilename(name),
		})
		htmlOutFile.Close()
		if err != nil {
			panic(err)
		}
	}
}

func getIn(txtFilename string) io.ReadCloser {
	in, err := os.Open(txtFilename)
	if err != nil {
		panic(err)
	}
	return in
}

func getOutExport(pathPrefix, txtBaseName string) io.WriteCloser {
	svgBaseName := toSVGFilename(txtBaseName)
	out, err := os.Create(pathPrefix + svgBaseName)
	if err != nil {
		panic(err)
	}
	return out
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

func stripSuffix(basename string) string {
	return strings.Split(basename,".")[0]
}
