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

	flag.StringVar(&inputFilename, "i", "", "Input filename (default stdin)")
	flag.StringVar(&outputFilename, "o", "", "Output filename (default stdout for SVG)")
	flag.StringVar(&format, "f", "svg", "Output format: svg (default: svg)")
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
		goat.BuildAndWriteSVG(input, output)
	}
}
