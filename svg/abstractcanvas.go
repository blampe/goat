package svg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log" // Tests may want to suppress log output.
	"os"
	"path/filepath"
	"strings"

	"github.com/blampe/goat/internal"
)

// X  Uncomment and populate this if more than one character should be reported as illegal.
//var illegalSet = MakeRuneSet(
//	'	',   // TAB
//)

// provide callbacks to specific Ascii or Unicode handlers.
type AbstractCanvas interface {
	GetCommon() *CanvasCommon
	WriteSVGBody(io.Writer, *Config)
	ShouldMoveToTextRunes(XyIndex) bool
}

// CanvasCommon represents the input diagram, after the first stage of parsing.
type CanvasCommon struct {
	// units of cells
	Width, Height int

	Data        map[XyIndex]rune
	TextRunes   map[XyIndex]rune
}

// 'ac.CanvasCommon().text' will contain begin and end marks for text styling
//
// XX  Output can be to any io.Writer, therefore no corresponding filename is
//     guaranteed accessible for logging of errors.
func WriteCanvas(config *Config, ac AbstractCanvas,
	includeDefaultCSS bool, colorsOnlyBytes string,
	cssInclude []internal.NamedReadSeeker,    // XX pass 'defaultCSS' in this way -- from ALL callers?   
	dst io.Writer) {

	mustPrintS := func(s string) {
		internal.MustFPrintf(dst, `%s`, s)
	}
	mustPrintS(ac.GetCommon().OpenSvgElement())

	// Include this first, so individual properties can be overridden.
	//
	// See:
	//   https://drafts.csswg.org/mediaqueries-5/#prefers-color-scheme
	//   https://developer.mozilla.org/en-US/docs/Web/SVG/Element/style
	//   https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute
	//
	//   X   Note that all elements are drawn by SVG not HTML, and the guidance here about
	//       the CSS "color:" property is not valid:
	//           https://developer.mozilla.org/en-US/docs/Web/CSS/color
	//       Rather, the authority is:
	//           https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute
	if includeDefaultCSS {
		mustPrintS(
			newStyleElement(
				"source-independent defaults: shared by ASCII and UTF-8",
				defaultCSS + colorsOnlyBytes))
	}

	for _, cssR := range cssInclude {
		// XX  Generalize the read-in to indent four columns, for ease of eyeballing
		//     the eventual SVG output file.   
		bs, err := io.ReadAll(cssR)
		title := cssR.Name()
		if err != nil {
			err = fmt.Errorf("Error in %s: '%v'", title, err)
			log.Fatal(err)
		}
		mustPrintS(
			newStyleElement(title, string(bs)))
	}

	mustPrintS(OpenGElement())

	ac.WriteSVGBody(dst, config)

	mustPrintS(CloseGElement())
	mustPrintS(CloseSvgElement())
}

// text returns a slice of all text characters not belonging to part of the diagram.
// Must be stably sorted, to satisfy regression tests.
func (c *CanvasCommon) text() (textRunes []text) {
	for idx := range LeftRightMinor(c.Width, c.Height) {
		r, found := c.TextRunes[idx]
		if !found {
			continue
		}
		if r == 0 {
			log.Fatalf("found rune with value 0x%x", r)
		}
		textRunes = append(textRunes,
			text{
				Start: idx,
				r: r,
			})
	}
	return
}

// text corresponds to any rune not reserved for diagrams, or a
// rune ordinarily reserved but in this case stripped of its special meaning
// because surrounded by alphanumerics.
type text struct {
	Start	 XyIndex    // XX  non-intuitive name: start and end are the same cell
	r rune
}

func (t text) String() string {
	return fmt.Sprintf(`xy index: %#v, character: %q`, t.Start, string(t.r))
}

// Looks only at c.Data[], ignores c.TextRunes[].
// Returns the rune for ASCII Space i.e. ' ', in the event that map lookup fails.
//  XX  Name 'dataRuneAt()' would be more descriptive, but maybe too bulky.
func (c *CanvasCommon) RuneAt(i XyIndex) rune {
	if val, ok := c.Data[i]; ok {
		return val
	}
	return ' '
}

// NewCanvasCommon creates a fully-populated CanvasCommon according to GoAT-formatted text read from
// an io.Reader, consuming all bytes available.
func NewCanvasCommon(config *Config, in io.Reader) (c CanvasCommon) {
	scanner := bufio.NewScanner(in)
	split := bufio.ScanLines
	if config.LineFilter != nil {
		split = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			advance, token, err = bufio.ScanLines(data, atEOF)
			if token == nil || err != nil || atEOF {
				return
			}
			if ! config.LineFilter.Match(token) {
				token = nil
			}
			return
		}
	}
	// Set the split function for the scanning operation.
	scanner.Split(split)

	c, err := newCanvasCommon(scanner)
	if err != nil {
		fatalIoReaderError(err, in)
	}

	// XX ? Separate and promote to caller this phase of processing and
	//      the particular data structure it creates?
	c.TextRunes = make(map[XyIndex]rune)
	return
}

// XX  Refactor?  Move to files.go -- more general package?
func fatalIoReaderError(err error, in io.Reader) {
	fileName := filenameFromReader(in)
	if len(fileName) == 0 {
		fileName = "input diagram not read from a named file"
	}
	err = fmt.Errorf("Error in %s: '%v'",
		fileName, err)
	log.Fatal(err)
}
func filenameFromReader(in io.Reader) string {
	file, isFile := in.(*os.File)
	if isFile {
		return file.Name()
	}
	return ""
}

// Create and populate a 'data' map.
func newCanvasCommon(scanner *bufio.Scanner) (CanvasCommon, error) {
	data := make(map[XyIndex]rune)
	width := 0
	height := 0

	for scanner.Scan() {
		lineStr := scanner.Text()

		w := 0
		// X  Type of second value assigned from "for ... range" operator over a string is "rune".
		//               https://go.dev/ref/spec#For_statements
		//    But yet, counterintuitively, type of lineStr[_index_] is 'byte'.
		//               https://go.dev/ref/spec#String_types
		for _, r := range lineStr {
			//if r > 255 {
			//	fmt.Printf("linestr=\"%s\"\n", lineStr)
			//	fmt.Printf("r == 0x%x\n", r)
			//}
			if r == '	' {
				return CanvasCommon{}, fmt.Errorf("Found TAB at row %d, column %d",
					height+1, w)
			}
			i := XyIndex{w, height}
			data[i] = r
			w++
		}

		if w > width {
			width = w
		}
		height++
	}
	if height == 0 {
		// Return an error, for fuller error diagnostics to CLI user.
		return CanvasCommon{}, errors.New("input appears to be empty!")
	}
	return CanvasCommon{
		Data: data,
		Width: width,
		Height: height,
		TextRunes: nil,  // XX
	}, nil
}

// Move contents of every cell that appears, according to a tricky set of rules,
// to be "text", into a separate map: from data[] to textRunes[].
// So data[] and textRunes[] are an exact partitioning of the
// incoming grid-aligned runes.
func MoveToText(ac AbstractCanvas) {
	cc := ac.GetCommon()
	for i := range LeftRightMinor(cc.Width, cc.Height) {
		if ac.ShouldMoveToTextRunes(i) {
			cc.TextRunes[i] = cc.RuneAt(i)	// cc.RuneAt() Reads from cc.Data[]
		}
	}
	for i := range cc.TextRunes {
		delete(cc.Data, i)
	}
}

func CanvasString(ac AbstractCanvas) string {
	var buffer bytes.Buffer
	cc := ac.GetCommon()

	for h := 0; h < cc.Height; h++ {
		for w := 0; w < cc.Width; w++ {
			idx := XyIndex{w, h}

			// Search 'text' map; if nothing there try the 'data' map.
			r, ok := cc.TextRunes[idx]
			if !ok {
				r = cc.RuneAt(idx)
			}

			_, err := buffer.WriteRune(r)
			if err != nil {
				continue
			}
		}

		err := buffer.WriteByte('\n')
		if err != nil {
			continue
		}
	}

	return buffer.String()
}

func ToSVGFilename(txtFilename string) string {
	return strings.TrimSuffix(txtFilename, filepath.Ext(txtFilename)) + ".svg"
}
