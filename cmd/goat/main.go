package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/blampe/goat"
)

func main() {
	log.SetFlags(0)

	var inputFilename string
	var outputFilename string
	var format string
	var svgColorLightScheme string
	var svgColorDarkScheme string

	flag.StringVar(&inputFilename, "i", "", "Input filename (default stdin)")
	flag.StringVar(&outputFilename, "o", "", "Output filename (default stdout for SVG)")
	flag.StringVar(&format, "f", "svg", "Output format: svg (default: svg)")
	flag.StringVar(&svgColorLightScheme, "sls", "#000000", `short for -svg-color-light-scheme`)
	flag.StringVar(&svgColorLightScheme, "svg-color-light-scheme", "#000000",
		`See help for -svg-color-dark-scheme`)
	flag.StringVar(&svgColorDarkScheme, "sds", "#FFFFFF", `short for -svg-color-dark-scheme`)
	flag.StringVar(&svgColorDarkScheme, "svg-color-dark-scheme", "#FFFFFF",
		`Goat's SVG output attempts to learn something about the background being
 drawn on top of by means of a CSS @media query, which returns a string.
 If the string is "dark", Goat draws with the color specified by
 this option; otherwise, Goat draws with the color specified by option
 -svg-color-light-scheme.

 See https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme
`)
	flag.BoolVar(&goat.HollowCircles, "hollowcircles", false,
		`If set, the letter 'o' draws a hollow circle, with strokes possibly extending
into it; otherwise, the circle is filled with a computed inverse of the foreground
drawing color.`)
	flag.Parse()

	format = strings.ToLower(format)
	if format != "svg" {
		log.Fatalf("unrecognized format: %s", format)
	}

	input := os.Stdin
	if inputFilename != "" {
		if _, err := os.Stat(inputFilename); os.IsNotExist(err) {
			log.Fatalf("input file not found: %s", inputFilename)
		}
		var err error
		input, err = os.Open(inputFilename)
		defer input.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	output := os.Stdout
	if outputFilename != "" {
		var err error
		output, err = os.Create(outputFilename)
		defer output.Close()
		if err != nil {
			log.Fatal(err)
		}
		// warn the user if he is writing to an extension different to the
		// file format
		ext := filepath.Ext(outputFilename)
		if fmt.Sprintf(".%s", format) != ext {
			log.Printf("Warning: writing to '%s' with extension '%s' and format %s", outputFilename, ext, strings.ToUpper(format))
		}
	} else {
		// check that we are not writing binary data to terminal
		fileInfo, _ := os.Stdout.Stat()
		isTerminal := (fileInfo.Mode() & os.ModeCharDevice) != 0
		if isTerminal && format != "svg" {
			log.Fatalf("refuse to write binary data to terminal: %s", format)
		}
	}

	switch format {
	case "svg":
		goat.BuildAndWriteSVG(input, output,
			svgColorLightScheme, svgColorDarkScheme)
	}
}
