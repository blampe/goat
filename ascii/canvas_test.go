package ascii

import (
	"bytes"
	"testing"

	"github.com/blampe/goat/svg"
	"github.com/google/go-cmp/cmp"
)

func AssertEqual(t *testing.T, x, y interface{}) {
	if ! cmp.Equal(x, y) {
		t.FailNow()
	}
}

func TestReadASCII(t *testing.T) {

	var buf bytes.Buffer

	buf.WriteString(" +-->\n")
	buf.WriteString(" | å\n")
	buf.WriteString(" +----->")

	canvas := NewCanvas(&svg.Config{}, &buf)

	AssertEqual(t, canvas.GetCommon().Width, 8)
	AssertEqual(t, canvas.GetCommon().Height, 3)

	buf.Reset()
	buf.WriteString(" +-->   \n")
	buf.WriteString(" | å    \n")
	buf.WriteString(" +----->\n")

	expected := buf.String()
	AssertEqual(t, expected, svg.CanvasString(canvas))
}
