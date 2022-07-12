// Copyright 2022 Donald Mullis. All rights reserved.

// XXXX Rename to 'goatdocdown'?   C.f.  https://github.com/robertkrimen/godocdown
// Command goatdoc transforms the output of `go doc -all` into Github-flavored Markdown.
//
// Go comments may contain Goat-format ASCII diagrams, each of which will
// be processed into an SVG image.
//
// If a package is implemented by more than one .go file, and more than one contain
// a package-level comment, 'go doc -all' includes all of them in its output.
// Order of inclusion appears to be alphabetical by file name.
// The alphabetical sensitivity may be worked around by creating a single-purpose
// file "doc.go", and allowing no per-package commentary in other files.
//
//  XX  An alternative implementation strategy would be to build upon Go's standard
//      library package https://pkg.go.dev/go/doc
//
//  XXX Weaknesses of current implementation: See warning to user in the Usage message.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/blampe/goat"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

const (
	svgColorLightSchemeDefault = "#000000"
	svgColorDarkSchemeDefault = "#FFFFFF"
)

var (
	svgFilesPrefix,
	svgColorLightScheme,
	svgColorDarkScheme string

	beginRegex,
	endRegex string
)

// Split input stream from 'go doc' into blocks, either goat-tagged or not.
//
//   Feed each goat-tagged block to 'goat.BuildAndWriteSVG()'
//      Write output to a separate *semantically-named* file NAME.svg.
//      Replace the entire goat-tagged block of the input stream
//      with either of:
//            <img link="./NAME.svg">
//      or, inline SVG, if known to be acceptable to the eventual Markdown renderer.
//
//   Scan each non-goat-tagged block for Golang-specific annotations, adding
//   github-flavored Markdown markup in the style of https://pkg.go.dev/std
func main() {
	flag.StringVar(&svgFilesPrefix, "svgfilesprefix", "goatdoc.",
		`Each SVG diagram produced by the goat library receives its
own file, prefixed with this string, followed by a serial number, followed
by ".svg".  The names of these files is known to the output Markdown file,
which specifies links to them.
`)
	flag.StringVar(&svgColorLightScheme, "sls", svgColorLightSchemeDefault,
		`short for -svg-color-light-scheme`)
	flag.StringVar(&svgColorLightScheme, "svg-color-light-scheme", svgColorLightSchemeDefault,
		`See help for command 'goat'`)
	flag.StringVar(&svgColorDarkScheme, "sds", svgColorDarkSchemeDefault,
		`short for -svg-color-dark-scheme`)
	flag.StringVar(&svgColorDarkScheme, "svg-color-dark-scheme", svgColorDarkSchemeDefault,
		`See help for command 'goat'`)
	flag.StringVar(&beginRegex, "goat-begin-re", `<goat>`,
		`UTF-8-art follows the input line matching this pattern.
The line itself is discarded`)
	flag.StringVar(&endRegex, "goat-end-re", `</goat>`,
		`UTF-8-art precedes the input line matching this pattern.
The line itself is discarded`)

	flag.Usage = func() {
		UsageDump()
		os.Exit(1)
	}
	flag.Parse()
	if !flag.Parsed() {
		log.Fatalln("flag.Parsed() == false")
	}

	scanner := bufio.NewScanner(os.Stdin)
	formatHeader(scanner)
	formatBody(scanner)
}

func writeUsage(out io.Writer, preamble, coda string) {
	fmt.Fprintf(out, "%s%s", preamble,
  `Usage:
        go doc -all | goatdoc >$(go list -f {{.Name}}).goatdoc.md
`)
	flag.PrintDefaults()
	fmt.Fprintf(out, "%s\n", coda)
}

// Comment preceding global function UsageDump().
func UsageDump() {
	writeUsage(os.Stderr, `
Be aware of the following limitations, consequences of all Go source passing first
through 'go doc all', before reaching 'goatdoc':
    - The first column of all lines to be included in a drawing must be a SPACE, to
      tell 'go doc' that the line is quoted "code", therefore not to be reflowed.
    - Any sequence of more than one blank line in a 'goat' block will be flattened
      to a single blank line (by 'go doc').

`, `
Each SVG file produced contains a CSS @media 'prefers-color-scheme' query; this
supports use within web pages that use a similar @media query to switch between
light and dark immediately upon the browser user demanding a different color
schema.

An SVG must perform its own @media query when it resides in a file of its own,
included at run time through an <img> element link, because it is then a
"replaced element", which cannot inherit any state from the HTML surrounding the
<img> element.
`)
}

type mapping struct {
	re *regexp.Regexp
	output string
}

var (
	h3 = mapping{
		re: regexp.MustCompile(`^(CONSTANTS|VARIABLES|FUNCTIONS|TYPES)`),
		output: "  ### ",
	}
)

func formatHeader(scanner *bufio.Scanner) {
	fmt.Printf("%s %s\n", h3.output, "Overview")

	var (
		packageImportLine = mapping{
			re: regexp.MustCompile(`^package .* // import`),
		}
		goatBegin = mapping{
			re: regexp.MustCompile(beginRegex),
		}
	)

	scanner.Scan()
	// Two cases, commands versus standard library packages.
	// Input from "go doc -all" writes out a line with importing info for the latter;
	// discard it.
	if line := scanner.Text(); packageImportLine.re.MatchString(line) {
		// Consume following blank line as well
		scanner.Scan()
	}

	// Unlike later comments, the initial, package-level overview section is not
	// indented, (as produced by 'go doc -all').
	for ; ; scanner.Scan() {
		line := scanner.Text()
		// Break at first section header e.g. "FUNCTIONS" or other.
		if h3.re.MatchString(line) {
			break
		}
		// Check for Goat diagram lines
		if goatBegin.re.MatchString(line) {
			formatGoat(scanner)
			continue
		}
		fmt.Println(line)
	}
	fmt.Println()
}

func formatGoat(scanner *bufio.Scanner) {
	var (
		buff bytes.Buffer
		goatEnd = mapping{
			re: regexp.MustCompile(endRegex),
		}
	)
	for scanner.Scan() {
		line := scanner.Text()
		if goatEnd.re.MatchString(line) {
			var out *os.File
			if svgFilesPrefix != "" {
				svgFilename := svgFilesPrefix + ".svg"
				fmt.Printf("![](./%s)\n", svgFilename)

				var err error
				out, err = os.Create(svgFilename)
				if err != nil {
					panic(err)
				}
			} else {
				out = os.Stdout
			}
			goat.BuildAndWriteSVG(
				&buff, out,
				svgColorLightScheme, svgColorDarkScheme)
			return
		}

		// append to buffer of entire drawing
		buff.WriteString(line)
		buff.WriteRune('\n')
	}
	panic("Found opening " + beginRegex + "; failed to find closing " + endRegex)
}

func formatBody(scanner *bufio.Scanner) {
	var (
		bodyText = mapping{
			re: regexp.MustCompile(`^    [A-Za-z0-9]`),
		}
		perMethod = mapping{
			re: regexp.MustCompile(`^func [(](?:[^ ]+ )([^)]+)[)] ([A-Za-z0-9]+)`),
			//                               ^^^^^^^^^^^^^^^ receiver
			output: "####",
		}
		perGlobal = mapping{
			re: regexp.MustCompile(`^(func|type|const) [A-Za-z0-9]+`),
			output: "####",
		}
	)

	//var thematicBreakRegex = regexp.MustCompile(`^type `)
	//const thematicBreak = "***"

	for {
		line := scanner.Text()
		//if thematicBreakRegex.MatchString(line) {
		//	fmt.Println(thematicBreak)
		//}

		if len(line) == 0 {
			fmt.Println()
		} else if h3.re.MatchString(line) {
			fmt.Printf("%s%s%s\n",
				h3.output, string(line[0]), strings.ToLower(line[1:]))
		} else if bodyText.re.MatchString(line) {
			// XX  Check for Goat diagram lines?

			// Discard the initial four spaces preferred by 'go doc', which
			// would trigger a code block from Markdown parsers.
			fmt.Println(strings.TrimLeft(line, " "))
		} else {
			if perMethod.re.MatchString(line) {
				// Strip out the name of receivers e.g.
				//       "xyz " from "(xyz *XYZType)"
				submatches := perMethod.re.FindStringSubmatch(line)
				fmt.Printf("%s func (%s) %s\n\n",
					perMethod.output,
					submatches[1], submatches[2])
			} else if perGlobal.re.MatchString(line) {
				fmt.Printf("%s %s\n\n",
					perGlobal.output,
					perGlobal.re.FindString(line))
			}
			fmt.Printf("      %s\n", line)
		}
		if !scanner.Scan() {
			break
		}
	}
}
