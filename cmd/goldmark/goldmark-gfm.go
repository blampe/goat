/*
   Produce GitHub-Flavored Markdown, for local proofing of README.md
*/
package main

import (
//	"bytes"
	"errors"
	"io"
	"log"
	"os"

	"github.com/yuin/goldmark"

//	"github.com/abhinav/goldmark-anchor"
// XXX  Elicits complaint:
//          go: downloading github.com/abhinav/goldmark-anchor v0.2.0
//          go: github.com/abhinav/goldmark-anchor@upgrade (v0.2.0) requires github.com/abhinav/goldmark-anchor@v0.2.0: parsing go.mod:
//          	module declares its path as: go.abhg.dev/goldmark/anchor
//          	        but was required as: github.com/abhinav/goldmark-anchor

	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"

	//"github.com/yuin/goldmark/renderer/html"

)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	//parser := goldmark.DefaultParser()
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			extension.GFM,
			//&anchor.Extender{},
		),
		//goldmark.WithRendererOptions(
		//	html.WithHardWraps(),
		//	html.WithXHTML(),
		//),
	)

	inFile := os.Stdin
	bS, err := io.ReadAll(inFile)
	if err != nil {
		panic(err)
	}
	if len(bS) == 0 {
		panic(errors.New("attempt to read file of zero length"))
	}

	if err := md.Convert(bS, os.Stdout); err != nil {
		panic(err)
	}
}
