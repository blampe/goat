// Copyright 2022 Donald Mullis. All rights reserved.

// XXXX Rename to 'goatdocdown'?
// Command goatdoc transforms the output of `go doc -all` into Github-flavored Markdown.
//
// XXXX Package-level overview comments
// Go comments may contain Goat-format ASCII diagrams, each of which will
// be processed into an SVG image.
//
// If a package is implemented by more than one .go file, and more than one contain
// a package-level comment, all are included in the output of 'go doc -all'.
//  X Order of inclusion appears to be alphabetical by file name.
// An alternative is to create a file "doc.go" centralizing all per-package commentary.
//
//  XX  An alternative implementation strategy would be to build upon Go's standard
//      library package https://pkg.go.dev/go/doc
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

var usageAbstract = `
Each SVG file produced contains a CSS @media 'prefers-color-scheme' query;
this is to support use within web pages that use a similar query to
switch between light and dark immediately upon the browser user demanding a different
color schema.
`
func writeUsage(out io.Writer, preamble string) {
	fmt.Fprintf(out, "%s%s", preamble,
  `Usage:
        go doc -all | goatdoc >$(go list -f {{.Name}}).goatdoc.md
`)
	flag.PrintDefaults()
	fmt.Fprintf(out, "%s\n", usageAbstract)
}

// Comment preceding global function UsageDump().
func UsageDump() {
	writeUsage(os.Stderr, "")
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
		goatStart = mapping{
			re: regexp.MustCompile(`^<goat>`),
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
		if goatStart.re.MatchString(line) {
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
			re: regexp.MustCompile(`</goat>`),
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
	panic("Found opening <goat>; failed to find closing </goat>.")
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
