// Copyright 2022 Donald Mullis. All rights reserved.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"regexp"
	"strings"
)

func UsageMarkdown() {
	var bytes strings.Builder
	flag.CommandLine.SetOutput(&bytes)

	writeUsage(&bytes, `<!-- Automatically generated Markdown, do not edit -->
 <style type="text/css">
 h3 {margin-block-end: -0.5em;}
 h4 {margin-block-end: -0.5em;}
 code {font-size: larger;}
 </style>
`)
	indentedTextToMarkdown(bytes)
}

var column1Regex = regexp.MustCompile(`^[A-Z]`)
const column1AtxHeading = "  ### "

var column3Regex = regexp.MustCompile(`^  [^ ]`)
const column3AtxHeading = "  #### "
// https://github.github.com/gfm/#atx-headings

// writes to stdout
func indentedTextToMarkdown(bytes strings.Builder) {
	scanner := bufio.NewScanner(strings.NewReader(bytes.String()))
	for scanner.Scan() {
		line := scanner.Text()
		if column1Regex.MatchString(line) {
			line = column1AtxHeading + line
		} else if column3Regex.MatchString(line) {
			line = column3AtxHeading + line
		}
		fmt.Println(line)
	}
}
