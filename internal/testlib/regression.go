package testlib

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/blampe/goat"
	"github.com/blampe/goat/internal"
	"github.com/blampe/goat/svg"
)

const (
	ExamplesDir = "examples"

	// Make these colors distinctive to this test, to emphasize upon
	// inspection that they are deliberately set and not simply inherited.
	svgColorLightScheme = "#212"
	svgColorDarkScheme =  "#FEF"

	// Text foreground colors chosen to contrast with graphics, to help expose any
	// misidentified characters.
	lightModeText = "#900"
	darkModeText = "#FBB"
)

// X  To dump usage message:
//     $ go test -v -args -h
var (
	write = flag.Bool("write",
		false, "write reference SVG output files")

	// Begin the directory name with '_' to hide from git.
	//  XX  ./_examples_new/ will by default appear under the CWD of the shell starting `go test ...`  
	svgDeltaDir = flag.String("svg-delta-dir", "_examples_new",
		`Directory to be filled with a delta-image file for each
newly-generated SVG that does not match those in ` + ExamplesDir)
)

// XX  promote to goat/ ?
type newCanvasFunc func(*svg.Config, io.Reader) svg.AbstractCanvas

func Regression(t *testing.T,
	reservedSet goat.RuneSet,
	newCanvas newCanvasFunc) {

	// XX  This sweeps up ~every~ *.txt file in examples/
	txtPaths, err := filepath.Glob(filepath.Join(ExamplesDir, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}

	// XX  How to produce this with less code?
	baseNames := make([]string, len(txtPaths))
	for i := range txtPaths {
		baseName, found := strings.CutPrefix(txtPaths[i], ExamplesDir+"/")
		if !found {
			panic("Could not cut prefix from pathname.")
		}
		baseNames[i] = baseName
	}

	// XX  DRY with svg.defaultCSS
	cssBytes := fmt.Appendf([]byte{},
`%s
    svg {
        background-color: %s
    }
    text {
        color: %s
    }
    @media (prefers-color-scheme: dark) {
        svg {
            background-color: %s
        }
        text {
            color: %s
        }
    }
`,
		svg.ColorsOnlyCssFileContent(svgColorLightScheme, svgColorDarkScheme),
		svgColorDarkScheme,
		lightModeText,
		svgColorLightScheme,
		darkModeText,
	)

	markBindingMap := make(svg.MarkBindingMap)
	err = svg.ParseCss(markBindingMap, cssBytes)
	if err != nil {
		t.Fatalf(`
    Could not parse embedded CSS: "%v"`, err)
	}
	config, err := svg.NewConfig(reservedSet, markBindingMap)
	config.LineFilter = noHashCommentLine_re
	if err != nil {
		t.Fatalf(`
    Aborting: "%v"`, err)
	}
	if *write {
		writeExamples(t, &config, newCanvas, string(cssBytes), ExamplesDir, ExamplesDir, baseNames)
	} else {
		t.Logf("Verifying equality of current SVG with examples/ references.\n")
		verifyExamples(t, &config, newCanvas, string(cssBytes), ExamplesDir, baseNames)
	}
}

// Allow sh-style comment lines in the regression test diagrams. 
// X  Will match a blank line -- no non-# printable character required.
var noHashCommentLine_re = regexp.MustCompile(`^(?:[^#].*)?$`)

func writeExamples(t *testing.T,
	config *svg.Config, newCanvas newCanvasFunc, cssBytes string,
	inDir, outDir string, baseNames []string) {

	for _, name := range baseNames {
		in := internal.MustOpen(inDir + "/" + name)
		ac := newCanvas(config, in)
		out := mustCreate(outDir + "/" + name)
		t.Logf("Writing new SVG to %s", out.Name())
		WriteCanvasNoCssFiles(config, ac, cssBytes, out)
		in.Close()
		out.Close()
	}
}

// XX  specific to creation of SVG output files
func mustCreate(txtFilename string) *os.File {
	return internal.MustCreate(svg.ToSVGFilename(txtFilename))
}

func verifyExamples(t *testing.T,
	config *svg.Config, newCanvas newCanvasFunc, cssBytes string,
	ExamplesDir string, baseNames []string) {

	var failures []string
	for _, name := range baseNames {
		t.Logf("Reading test case file: %s", name)
		in := internal.MustOpen(ExamplesDir + "/" + name)
		ac := newCanvas(config, in)
		buff := &bytes.Buffer{}
		WriteCanvasNoCssFiles(config, ac, cssBytes, buff)
		in.Close()
		if nil != CompareSVG(t, buff, ExamplesDir, name) {
			failures = append(failures, name)
		}

	}
	if len(failures) > 0 {
		t.Logf(`Failed to verify contents of %d .svg files`,
			len(failures))
		err := os.Mkdir(*svgDeltaDir, 0770)
		if err != nil {
//			cwd, _ := os.Getwd()
			t.Fatalf(`
    Aborting attempt to write out visual difference files: "%v"
    Consider 'rm -r %s'`,
				err,
				//cwd,
				*svgDeltaDir)
		}

		// X  Set foreground colors to contrast with usual, for ease of visual diffing.
		cssStr := svg.ColorsOnlyCssFileContent("#000088", "#88CCFF")
		writeExamples(t, config, newCanvas, cssStr, ExamplesDir, *svgDeltaDir, failures)
		writeDeltaHTML(t, "../" + ExamplesDir, *svgDeltaDir, failures)
		t.FailNow()
	}
}

func WriteCanvasNoCssFiles(config *svg.Config, ac svg.AbstractCanvas, cssStr string, dst io.Writer) {
	svg.WriteCanvas(config, ac,
		true, // includeDefaultCSS
		cssStr, []internal.NamedReadSeeker{}, dst)
}

func CompareSVG(t *testing.T, buff *bytes.Buffer, ExamplesDir string, baseName string) error {
	fileName := ExamplesDir + "/" + baseName
	golden, err := getOutString(fileName)  // XX  rename 'golden' to 'expected'?
	if err != nil {
		t.Log(err)
	}
	if newStr := buff.String(); newStr != golden {
		// Skip complaint if the modification timestamp of the .txt file
		// source is fresher than that of the .svg?
		//   => NO, Any .txt difference might be an editing mistake.

		t.Logf("Content mismatch for %s. Length was %d, expected %d",
			svg.ToSVGFilename(fileName), buff.Len(), len(golden))
		for i:=0; i<min(len(golden), len(newStr)); i++ {
			if newStr[i] != golden[i] {
				t.Logf("Differing runes at offset %d: new='%#v' reference='%#v'\n",
					i, newStr[i], golden[i])
				break
			}
		}
		t.Logf("Generated contents do not match existing %s",
			svg.ToSVGFilename(fileName))
		return errors.New("Generated contents do not match existing")
	} else {
		//if testing.Verbose() {
		//	t.Logf("Existing and generated contents match %s\n",
		//		goat.ToSVGFilename(fileName))
		//}
	}
	return nil
}

// See https://developer.mozilla.org/en-US/docs/Web/CSS/blend-mode#example_using_difference
func writeDeltaHTML(t *testing.T, ExamplesDir, deltaDir string, baseNames []string) {
	t.Logf("Writing new SVG and HTML visual difference files into %s/", deltaDir)

	tmpl := template.Must(template.New("_ignored_").Parse(`
<style type="text/css">
.blended-images {
    height: 100%; /* XX  How to replace with pixel height of the pair of SVGs? */
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
		htmlOutFile := internal.MustCreate(deltaDir + "/" + htmlOutName)
		err := tmpl.Execute(htmlOutFile, map[string]string{
			"ExamplesDir": ExamplesDir,
			"DeltaDir": ".",
			"SvgBaseName": svg.ToSVGFilename(name),
		})
		htmlOutFile.Close()
		if err != nil {
			panic(err)
		}
	}
}

func getOutString(txtFilename string) (string, error) {
	b, err := ioutil.ReadFile(svg.ToSVGFilename(txtFilename))
	if err != nil {
		// XX  Simply panic rather than return an error?
		return "", err
	}
	// XX  Why are there RETURN characters in contents of the .SVG files?
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return string(b), nil
}

func stripSuffix(basename string) string {
	return strings.Split(basename,".")[0]
}

