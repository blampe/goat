package utf8_test

import (
	"flag"
	"testing"

	"github.com/blampe/goat/internal/testlib"

	"github.com/blampe/goat/utf8"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !flag.Parsed() {
		panic("flag.Parsed() == false")
	}
	m.Run()
}

func TestExamples(t *testing.T) {
	testlib.Regression(t, utf8.ReservedSet, utf8.NewCanvas)
}
