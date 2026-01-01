package internal

// General library functions useful to non-server, exit-on-error CLI programs.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type NamedReadSeeker interface {
	io.ReadSeeker

	// bytes.Reader lacks such a field
	Name() string
}

type NamedBytesReader struct {
	*bytes.Reader

	// bytes.Reader lacks such a field
	name string
}

func NewNamedBytesReader(b []byte, name string) NamedBytesReader {
	return NamedBytesReader{
		Reader: bytes.NewReader(b),
		name: name,
	}
}

func (nr NamedBytesReader) Name() string {
	return nr.name
}

func MustOpen(filename string) *os.File {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatalf("input file not found: %s", filename)
	}
	in, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	return in
}

func MustCreate(filename string) *os.File {
	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func fileError(err error, fr NamedReadSeeker) {
	where := fmt.Sprintf("on file name %s", fr.Name())
	log.Output(2, where + err.Error())
	log.Fatalln("exiting")
}

func ReadFileAll(fr NamedReadSeeker) []byte {       // XX  called only from parseargs.go
	bS, err := io.ReadAll(fr)
	if err != nil {
		fileError(err, fr)
	}
	if len(bS) == 0 {
		fileError(errors.New("attempt to read file of zero length"), fr)
	}
	return bS
}

func MustFPrintf(out io.Writer, format string, args ...interface{}) {
	_, err := fmt.Fprintf(out, format, args...)
	if err != nil {
		log.Fatal(err)
	}
}
