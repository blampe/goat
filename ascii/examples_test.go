package ascii_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/blampe/goat/internal"
	"github.com/blampe/goat/svg"
	"github.com/blampe/goat/internal/testlib"

	"github.com/blampe/goat/ascii"
)

// Catch iterations over maps.
// X  TXT source file suite is limited to a single file -- "circuits.txt"
func TestExampleStableOutput(t *testing.T) {
	var previous string
	for i := 0; i < 3; i++ {
		in, err := os.Open(filepath.Join(testlib.ExamplesDir, "circuits.txt"))
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		// XX  Better to test also the API functions that generates this non-trivially.
		config := svg.Config{}
		ac := ascii.NewCanvas(&config, in)
		testlib.WriteCanvasNoCssFiles( &config, ac, "", &out)
		in.Close()
		if i > 0 && previous != out.String() {
			t.FailNow()
		}
		previous = out.String()
		//fmt.Println(previous)
	}
}

func BenchmarkComplicated(b *testing.B) {
	in := internal.MustOpen(filepath.FromSlash("examples/complicated.txt"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac := ascii.NewCanvas(&svg.Config{}, in)
		testlib.WriteCanvasNoCssFiles( &svg.Config{}, ac, "", io.Discard)
	}
	in.Close()
}
