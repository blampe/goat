package main

import (
	"log"
	"os"

	goat "github.com/bep/goat/src"

	"gopkg.in/alecthomas/kingpin.v2"
)

var fileName = kingpin.Arg(
	"file",
	"Path to a file containing an ASCII diagram.",
).Required().String()

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatal(err)
	}

	goat.ASCIItoSVG(file, os.Stdout)
}
