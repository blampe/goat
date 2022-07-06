package main

import (
	"flag"
	"log"
	"os"

	"github.com/blampe/goat"
)

func init() {
	log.SetFlags(/*log.Ldate |*/ log.Ltime | log.Lshortfile)
}

func main() {
	var (
		inputFilename,
		outputFilename,
		svgColorLightScheme,
		svgColorDarkScheme string
	)

	flag.StringVar(&inputFilename, "i", "", "Input filename (default stdin)")
	flag.StringVar(&outputFilename, "o", "", "Output filename (default stdout for SVG)")
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
	}
	goat.BuildAndWriteSVG(input, output,
		svgColorLightScheme, svgColorDarkScheme)
}
