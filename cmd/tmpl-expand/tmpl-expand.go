// Copyright 2025 Donald Mullis. All rights reserved.

// See `tmpl-expand -h` for abstract.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
//	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/blampe/goat/internal"
)

type (
	KvpArg struct {
		Key    string
		Value string
	}
	TemplateContext struct {
		//    https://golang.org/pkg/text/template/#hdr-Arguments
		tmpl *template.Template
		substitutionsMap map[string]string
	}
)

// General args
var (
	writeMarkdown = flag.Bool("markdown", false,
		`Reformat -help usage message into Github-flavored Markdown`)

	exitStatus int
)

func init() {
	log.SetFlags(log.Lshortfile)

	pwd, _ := os.Getwd()
	exe, _ := os.Executable()
	prefixStr := fmt.Sprintf(`
Executable %s in CWD %s:
`,
		exe, pwd)
	log.SetPrefix(prefixStr)
}

func main() {
	flag.Usage = func() {
		UsageDump()
		os.Exit(1)
	}
	flag.Parse()
	if !flag.Parsed() {
		log.Fatalln("flag.Parsed() == false")
	}

	if *writeMarkdown {
		UsageMarkdown()
		return
	}

	kvpArgs, defFileNameArgs := scanForKVArgs(flag.Args())
	for _, filename := range defFileNameArgs {
		kvpArg := scanValueFile(filename)
		kvpArgs[kvpArg.Key] = kvpArg.Value

	}
	templateText := getTemplate(os.Stdin)
	ExpandTemplate(kvpArgs, templateText)
	os.Exit(exitStatus)
}

var usageAbstract = `
  Key=Value
   Sh-style name=value definition string pairs.
   The Key name string must be valid as a Go map Key acceptable to Go's
   template package -- see https://pkg.go.dev/text/template

  ValueFilePath
   Alternate means of specifying a Key=Value pair, allowing long Values.
   The arg must be a valid path to a file, the entire contents of which
   will be used as the Value.
   The 'Key' is derived from the pathname itself by mapping
   any non-alphanumeric character to "_".
   The result is therefore legal as the name of a template Key.

  TemplateFile
   A stream read from stdin containing references to the 'Key' side of
   the above pairs, embedded within Go template library constructs.

  ExpansionFile
   Written to stdout, the expansion of the input template read from stdin.

---
Example:

      $ echo >/tmp/valueFile.txt '
      .      +-------+
      .      | a box |
      .      +-------+
      . '
      $
      $ echo '
      .     A sentence referencing Key "boxShape" with Value "{{.boxShape}}", read
      .     from the command line.
      .
      .     An introductory clause followed by a multi-line block of text,
      .     read from a file:
      .       {{.valueFile_txt}}
      . ' |
        tmpl-expand boxShape='RECTANGULAR' /tmp/valueFile.txt

Result:
      .     A sentence referencing Key boxShape with Value RECTANGULAR, read
      .     from the command line.
      .
      .     An introductory clause followed by a multi-line block of text,
      .     read from a file:
      .
      .      +-------+
      .      | a box |
      .      +-------+
`

func writeUsage(out io.Writer, premable string) {
	fmt.Fprintf(out, "%s%s", premable,
  `Usage:
    tmpl-expand [-markdown] [ Key=Value | ValueFilePath ] ... <TemplateFile >ExpansionFile

`)
	flag.PrintDefaults()
	fmt.Fprintf(out, "%s\n", usageAbstract)
}

func UsageDump() {
	writeUsage(os.Stderr, "")
}

func scanForKVArgs(args []string) (
	kvpArgs map[string]string, filenameArgs []string) {
	kvpArgs = make(map[string]string)
	for _, arg := range args {
		kvp := strings.Split(arg, "=")
		if len(kvp) != 2 {
			// Any duplicated file names would be harmless -> ignored.
			filenameArgs = append(filenameArgs, kvp[0])
			continue
		}
		newKvpArg := newKVPair(kvp)

		// Search earlier Keys for duplicates.
		//   XX  N^2 in number of Keys -- use a map instead.
		for k := range kvpArgs {
			if k == newKvpArg.Key {
				log.Printf("Duplicate key specified: '%v', '%v'", kvp, newKvpArg)
				exitStatus = 1
			}
		}
		kvpArgs[newKvpArg.Key] = newKvpArg.Value
	}
	return
}

func newKVPair(newKvp []string) KvpArg {
	vetKVstring(newKvp)
	return KvpArg{
		Key:   newKvp[0],
		Value: newKvp[1],
	}
}

func vetKVstring(kv []string) {
	reportFatal := func(format string) {
		// X X   Caller disappears from stack, apparently due to inlining, despite
		//       disabling Go optimizer
		//caller := func(howHigh int) string {
		//	pc, file, line, ok := runtime.Caller(howHigh)
		//	_ = pc
		//	if !ok {
		//		return ""
		//	}
		//	baseFileName := file[strings.LastIndex(file, "/")+1:]
		//	return baseFileName + ":" + strconv.Itoa(line)
		//}
		log.Printf(format, kv)
		log.Fatalln("FATAL")
	}
	if len(kv[0]) <= 0 {
		reportFatal("Key side of Key=Value pair empty: %#v\n")
	}
	if len(kv[1]) <= 0 {
		reportFatal("Value side of Key=Value pair empty: %#v\n")
	}
}

var alnumOnlyRE = regexp.MustCompile(`[^a-zA-Z0-9]`)

func scanValueFile(keyPath string) KvpArg {
	valueFile, err := os.Open(keyPath)
	if err != nil {
		log.Fatalln(err)
	}
	bytes, err := io.ReadAll(valueFile)
	if err != nil {
		log.Fatalln(err)
	}

	//basename := path.Base(keyPath)
	return KvpArg{
		//Key:   alnumOnlyRE.ReplaceAllLiteralString(basename, "_"),
		Key:   alnumOnlyRE.ReplaceAllLiteralString(keyPath, "_"),
		Value: string(bytes),
	}
}

func getTemplate(infile *os.File) string {
	var err error
	var stat os.FileInfo
	stat, err = infile.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	templateText := make([]byte, stat.Size())
	var nRead int
	templateText, err = io.ReadAll(infile)
	nRead = len(templateText)
	if nRead <= 0 {
		log.Fatalf("os.Read returned %d bytes", nRead)
	}
	if err = infile.Close(); err != nil {
		log.Fatalf("Could not close %v, err=%v", infile, err)
	}
	return string(templateText)
}

func openAndReadFile(filePath string) string {
	file := internal.MustOpen(filePath)
	bytes := internal.ReadFileAll(file)
	return string(bytes)
}

func ExpandTemplate(kvpArgs map[string]string, templateText string) {
	var funcMap = template.FuncMap{
		"include": openAndReadFile,
		// "includeIndented": openReadAndIndentFile
	}

	ctx := TemplateContext{
		substitutionsMap: kvpArgs,
	}

	var err error
	ctx.tmpl, err = template.New(""/*template name*/).Option("missingkey=error").
		Funcs(funcMap).Parse(templateText)
	if err != nil {
		log.Fatalf(`
Error "%v" on this text:

"%s
`,
			err, templateText)
	}
	ctx.writeFile()
}

func (ctx *TemplateContext) writeFile() {
	if err := ctx.tmpl.Execute(os.Stdout, ctx.substitutionsMap); err != nil {
		fmt.Fprintf(os.Stderr, "Template.Execute(outfile, map) returned  err=\n   %v\n",
			err)
		fmt.Fprintf(os.Stderr, "Contents of failing map:\n%s", ctx.formatMap())
		exitStatus = 1
	}
	if err := os.Stdout.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

// Sort the output, for deterministic comparisons of build failures.
func (ctx *TemplateContext) formatMap() (out string) {
	alphaSortMap(ctx.substitutionsMap,
		func(s string) {
			out += fmt.Sprintf("   % 20s\n", s)
		})
	alphaSortMap(ctx.substitutionsMap,
		func(s string) {
			v := ctx.substitutionsMap[s]
			const TRIM = 80
			if len(v) > TRIM {
				v = v[:TRIM] + "..."
			}
			out += fmt.Sprintf("   % 20s '%v'\n\n", s, v)
		})
	return
}

func alphaSortMap(m map[string]string, next func(s string)) {
	var h sort.StringSlice
	for k, _ := range m {
		h = append(h, k)
	}
	h.Sort()
	for _, s := range h {
		next(s)
	}
}
