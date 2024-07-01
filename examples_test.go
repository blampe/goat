package goat

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// XX  TXT source file suite is limited to a single file -- "circuits.txt"
func TestExampleStableOutput(t *testing.T) {
	var previous string
	for i := 0; i < 3; i++ {
		in, err := os.Open(filepath.Join(examplesDir, "circuits.txt"))
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		BuildAndWriteSVG(in, &out, "black", "white")
		in.Close()
		if i > 0 && previous != out.String() {
			t.FailNow()
		}
		previous = out.String()

	}
}

func BenchmarkComplicated(b *testing.B) {
	in := getIn(filepath.FromSlash("examples/complicated.txt"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildAndWriteSVG(in, io.Discard, "black", "white")
	}
	in.Close()
}
