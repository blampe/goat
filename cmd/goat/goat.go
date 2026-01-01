/*
The command 'goat' provides CLI access to SVG images produced from text-art.
*/
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"regexp"

	"github.com/blampe/goat"
	"github.com/blampe/goat/ascii"
	"github.com/blampe/goat/css"
	"github.com/blampe/goat/svg"
	"github.com/blampe/goat/utf8"
)

func init() {
	log.SetFlags(/*log.Ldate |*/ log.Ltime | log.Lshortfile)

	pwd, _ := os.Getwd()
	exe, _ := os.Executable()
	prefixStr := fmt.Sprintf(`
Executable %s in CWD %s:
`,
		exe, pwd)
	log.SetPrefix(prefixStr)
}

func main() {
	args, cssInclude, markBindingMap := ParseFlags()
	colorsOnlyBytes := svg.ColorsOnlyCssFileContent(
			args.SvgColorLightScheme,
			args.SvgColorDarkScheme)

	if args.listEmbedded {
		dumpNames(css.FileSystem)
		return
	}

	input, output := OpenIO(&args)
	// XX  Necessary if a call to os.Exit() before draining of output buffer is possible.
	//	defer output.Close()   

	var (
		config svg.Config
		canvas svg.AbstractCanvas
	)
	// XX  Simplify by following example of testlib.newCanvasFunc ?
	if args.Utf8 {
		config = newConfig(&args, utf8.ReservedSet, markBindingMap)
		canvas = utf8.NewCanvas(&config, input)
	} else {
		config = newConfig(&args, ascii.ReservedSet, markBindingMap)
		canvas = ascii.NewCanvas(&config, input)
	}
	svg.WriteCanvas(&config, canvas,
		args.IncludeDefaultCSS, colorsOnlyBytes, cssInclude, output)
}

func dumpNames(efs embed.FS) {
	mustRD := func(path string) (ent []fs.DirEntry) {
		ent, err := efs.ReadDir(path)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	entries := mustRD(".")
	for _, dir := range entries {
		fileEntries := mustRD(dir.Name())
		for _, file := range fileEntries {
			fmt.Fprintf(os.Stderr, "%s:%s/%s\n",
				EmbedPrefix, dir.Name(), file.Name())
		}
	}
}

func newConfig(args *Args, reservedSet goat.RuneSet, markBindingMap svg.MarkBindingMap) (
	config svg.Config) {

	config, err := svg.NewConfig(reservedSet, markBindingMap)
	if err != nil {
		log.Fatal(err)
	}

	config.LineFilter = regexp.MustCompile(args.LineFilterRegexpString)
	return
}
