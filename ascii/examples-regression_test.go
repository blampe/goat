package ascii_test

import (
	"flag"
	"testing"

	"github.com/blampe/goat/internal/testlib"

	"github.com/blampe/goat/ascii"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !flag.Parsed() {
		panic("flag.Parsed() == false")
	}
	m.Run()
}

func TestExamples(t *testing.T) {
	testlib.Regression(t, ascii.ReservedSet, ascii.NewCanvas)
}
