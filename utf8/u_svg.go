package utf8

import (
	"io"

	"github.com/blampe/goat/internal"
	"github.com/blampe/goat/svg"
)

type smallTriangle struct {
	Start	     svg.XyIndex
	Orientation  svg.Orientation
}

// Draw a solid triangle as an SVG polygon element.
func (t smallTriangle) Draw(out io.Writer) {
	x, y := float32(t.Start.AsPixel().X), float32(t.Start.AsPixel().Y)
	r := 0.0

	// Coordinate values below are effective verbatim only for O_E, an isosceles
	// triangle "pointing" rightward; rotation to the desired final direction
	// is a post-process.
	//  XX  Parameterize the size?
	half := float32(1.5) // 4
	x0 := x + 2*half   // tip of the arrowhead
	y0 := y
	x1 := x - half
	y1 := y - 0.35*4*half
	x2 := x - half
	y2 := y + 0.35*4*half

	switch t.Orientation {
	case svg.O_E:
		r = 0
	case svg.O_W:
		r = 180
	case svg.O_S:
		r = 90
	case svg.O_N:
		r = 270
	}
	switch t.Orientation {
	case svg.O_S:
		fallthrough
	case svg.O_N:
		// advance
		x0 += 4
		x1 += 4
		x2 += 4
	}

	// <polygon> inherits both 'fill' and 'stroke' attributes from parents.
	internal.MustFPrintf(out, svg.PolygonPrintFmt,
		x0, y0,
		x1, y1,
		x2, y2,
		r,
		x, y)
}
