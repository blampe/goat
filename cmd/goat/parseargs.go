package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/blampe/goat/css"
	"github.com/blampe/goat/internal"
	"github.com/blampe/goat/svg"
)

const (
	EmbedPrefix = "embed"
)

type Args struct {
	listEmbedded, Utf8, IncludeDefaultCSS bool

	inputFilename,
	outputFilename,
	ioPathname,

	LineFilterRegexpString,

	// Default for all color fill, from command line.
	SvgColorLightScheme, SvgColorDarkScheme string
}

func ParseFlags() (
	args Args,
	cssInclude []internal.NamedReadSeeker,
	markBindingMap svg.MarkBindingMap) {

	const (
		white = "#FFF"
		black = "#000"
	)
	flag.BoolVar(&args.listEmbedded, "list-embedded", false,
		`Dump the names of CSS files embedded in the goat binary.
Extract the contents of one by naming it on the command line of another
invocation of goat, prefixed by "embed:"`)

	flag.BoolVar(&args.Utf8, "utf8", false,
		`Diagram input contains UTF-8 BOX characters.
Goat treats only these as graphics; ASCII characters regarded by Markdeep as
coding for graphics are treated as ordinary text.`)

	// abort if this is 'false' and -sls or -sds has been specified
	flag.BoolVar(&args.IncludeDefaultCSS, "defaultcss", true,
		`Prefix any other CSS content with a baseline stylesheet supporting a single color
each for light or dark mode`)

	flag.StringVar(&args.inputFilename, "i", "", "Input filename (default: standard input)")
	flag.StringVar(&args.outputFilename, "o", "", "Output filename (default: standard output)")
	flag.StringVar(&args.ioPathname, "io", "",
		`Path terminating in a .txt basename: Output will be directed
to a new file with pathname same except for replacement of .txt with .svg`)

	flag.StringVar(&args.LineFilterRegexpString, "regexp", "",
		"Discard any input lines that fail to match this regular expression.")

	flag.StringVar(&args.SvgColorLightScheme, "sls", black, `short for -svg-color-light-scheme`)
	flag.StringVar(&args.SvgColorLightScheme, "svg-color-light-scheme", black,
		`See help for -svg-color-dark-scheme`)

	flag.StringVar(&args.SvgColorDarkScheme, "sds", white, `short for -svg-color-dark-scheme`)
	flag.StringVar(&args.SvgColorDarkScheme, "svg-color-dark-scheme", white,
		`Goat's SVG output attempts to learn something about the background being
drawn on top of by means of a CSS @media query, which returns a string.
If the string is "dark", Goat draws with the color specified by
this option; otherwise, Goat draws with the color specified by option
-svg-color-light-scheme.
See:
     https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme
     https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute
Default value of 'currentColor' as tested on Firefox is one suitable for
a "light-mode" display, therefore is dark.
`)
	flag.Usage = func() {
		clOutput := flag.CommandLine.Output()
		fmt.Fprintf(clOutput, `Usage: %[1]s [flags] [CSS-filename ...]

%[1]s conforms to the Unix standard for CLI "filter" commands: Read from standard input;
process the incoming bytes as directed by CLI arguments; write to standard output.

Input is a UTF-8 encoded byte stream; output is a single <svg> element.
%[1]s reads no configuration files.

`,
			os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(clOutput, `  CSS-filename ...
	Non-flag args are assumed to be paths to CSS files.
	Goat wraps the contents of each within a SVG <style> element
        and appends the element within the output SVG; therefore,
        properties in CSS files later on the command line may override
        those specified by earlier files.
`)
	}
	flag.Parse()

	if !args.IncludeDefaultCSS {
		cliColorSettingArgs := map[string]struct{}{
			"sls": {},
			"svg-color-light-scheme": {},
			"sds": {},
			"svg-color-dark-scheme": {},
		}
		flag.Visit(
			func (fl *flag.Flag) {
				_, found := cliColorSettingArgs[fl.Name]
				if found {
					log.Fatalf("-includeDefaultCSS==false, but color option -%s specified",
						fl.Name)
				}
			})
	}

	// Multiple CSS files may be specified.
	markBindingMap = make(svg.MarkBindingMap)
	for _, cssFilename := range flag.Args() {
		switch ext := path.Ext(cssFilename); ext {
		case ".css":
			fallthrough
		case ".CSS":
			var newNR internal.NamedReadSeeker
			var bytes []byte
			var err error
			if after, found := strings.CutPrefix(cssFilename, EmbedPrefix + ":"); found {
				bytes, err = css.FileSystem.ReadFile(after)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				osFile := internal.MustOpen(cssFilename)
				bytes = internal.ReadFileAll(osFile)
			}
			newNR = internal.NewNamedBytesReader(bytes, cssFilename)

			cssInclude = append(cssInclude, newNR)

			newCss := internal.ReadFileAll(newNR)

			//  https://cs.opensource.google/go/go/+/go1.25.4:src/os/file.go;l=303
			//   XX  excessively clever?  More simply, change callee args to []byte ?
			_, _ = newNR.Seek(0, 0)

			err = svg.ParseCss(markBindingMap, newCss)
			if err != nil {
				log.Fatalf(`
Could not parse file '%s',
   err = %v`,
					cssFilename, err)
			}
		default:
			log.Fatalf(`
Expected filename with suffix .css, found %s with extension %s`,
				cssFilename, ext)
		}
	}
	return
}

func OpenIO(args *Args) (input, output *os.File) {
	input = os.Stdin
	if len(args.inputFilename) > 0 {
		input = internal.MustOpen(args.inputFilename)
		if len(args.ioPathname) > 0 {
			log.Fatalf("options %s and %s are mutually exclusive", "-i", "-io")
		}
	}
	output = os.Stdout
	if len(args.outputFilename) > 0 {
		if len(args.ioPathname) > 0 {
			log.Fatalf("options %s and %s are mutually exclusive", "-o", "-io")
		}
		output = internal.MustCreate(args.outputFilename)
	}
	if len(args.ioPathname) > 0 {
		input = internal.MustOpen(args.ioPathname)
		before, found := strings.CutSuffix(args.ioPathname, ".txt")
		if !found {
			log.Fatalf("%s does not end in %s", args.ioPathname, ".txt")
		}
		outFilename := before + ".svg"
		output = internal.MustCreate(outFilename)
	}
	return
}
