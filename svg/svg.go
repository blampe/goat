package svg

import (
	"fmt"
	"io"

	"github.com/blampe/goat/internal"
)

func (cc *CanvasCommon) OpenSvgElement() string {
	return fmt.Sprintf(
`<svg xmlns="http://www.w3.org/2000/svg" version="1.1"
    width="%d" height="%d"
    viewBox="0 0 %d %d">
`,
		cc.widthScreen(), cc.heightScreen(),
		cc.widthScreen(), cc.heightScreen(),
	)
}

func (c *CanvasCommon) heightScreen() int {
	// " + 8 + 2", because any less results in clipping of any edge at the bottom of the drawing
	//    XX  Necessary because fragilely tuned to 'edgeToCenterY', below.
	return c.Height*CellHeight + 8 + 2
}

func (c *CanvasCommon) widthScreen() int {
	// XX  "c.Width + 1", fragilely tuned to 'edgeToCenterX', below.
	return (c.Width + 1) * CellWidth
}

func CloseSvgElement() string {
	return `</svg>
`
}

// We desire that pixel coordinate {0,0} should lie at the *center* of the 8x16
// "cell" at top-left corner of the enclosing SVG element, and that a
// visually-pleasing margin separate that cell from the visible top-left
// corner; the 'translate()' below accomplishes that.
//
// X The former 16-pixel Y-axis translation of the transform was more than necessary – there
// is an always-blank band at the top of SVG image.
func OpenGElement() string {
	edgeToCenterX, edgeToCenterY :=
		CellWidth, CellHeight*3/4

	return fmt.Sprintf(`
<g transform='translate(%d,%d)'>
`,
		edgeToCenterX, edgeToCenterY)
}

func CloseGElement() string {
	return `</g>
`
}

func WritePolyline(out io.Writer, start, stop Pixel) {
	internal.MustFPrintf(out, `    <polyline class="path" points="%d,%d %d,%d"/>
`,
		start.X, start.Y,
		stop.X, stop.Y,
	)
}

// Draw a solid triangle as an SVG polygon element.
func (t Triangle) Draw(out io.Writer) {
	// https://www.w3.org/TR/SVG/shapes.html#PolygonElement

	/*
		  +-----+-----+
		  |    /|\    |
		  |   / | \   |
		x +- / -+- \ -+
		  | /	|   \ |
		  |/	|    \|
		  +-----+-----+
			y
	*/

	x, y := float32(t.Start.AsPixel().X), float32(t.Start.AsPixel().Y)
	r := 0.0

	// Coordinate values below are effective verbatim only for O_E, an isosceles
	// triangle "pointing" rightward; rotation to the desired final direction
	// is a post-process.
	// If regarded as an arrow-point, it will be 12px long and
	// 0.35*16*2=11.2px in width.
	x0 := x + W
	y0 := y
	x1 := x - W/2
	y1 := y - 0.35*CellHeight
	x2 := x - W/2
	y2 := y + 0.35*CellHeight

	// 't.NeedsNudging' in all cases below means "put the tip of the arrowhead
	// on the boundary of the next cell".
	// For a horizontal arrowhead this means retreating so as not to intrude
	// into the next cell; for other orientations it necessitates advancing. XX  correct?      
	switch t.Orientation {
	case O_N:
		r = 270
		if t.NeedsNudging {
			x0 += W
			x1 += W
			x2 += W
		}
	case O_NE:
		r = 300
		x0 += W/2
		x1 += W/2
		x2 += W/2
		if t.NeedsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case O_NW:
		r = 240
		x0 += W/2
		x1 += W/2
		x2 += W/2
		if t.NeedsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case O_W:
		r = 180
		if t.NeedsNudging {
			x0 -= W
			x1 -= W
			x2 -= W
		}
	case O_E:
		r = 0
		if t.NeedsNudging {
			x0 -= W
			x1 -= W
			x2 -= W
		}
	case O_S:
		r = 90
		if t.NeedsNudging {
			x0 += W
			x1 += W
			x2 += W
		}
	case O_SW:
		r = 120
		x0 += W/2
		x1 += W/2
		x2 += W/2
		if t.NeedsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case O_SE:
		r = 60
		x0 += W/2
		x1 += W/2
		x2 += W/2
		if t.NeedsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	}

	// <polygon> inherits both 'fill' and 'stroke' attributes from parents.
	internal.MustFPrintf(out, PolygonPrintFmt,
		x0, y0,
		x1, y1,
		x2, y2,
		r,
		x, y)
}

const PolygonPrintFmt = `    <polygon points="%g,%g %g,%g %g,%g" transform="rotate(%g, %g, %g)" class="arrowhead"></polygon>
`

// Draw a solid circle as an SVG circle element.
func (ci *Circle) Draw(out io.Writer, circleRadius int) {
	var class string
	if ci.Bold {    // bad name?
		class = "filled"
	} else {
		class = "hollow"
	}
	pixel := ci.Start.AsPixel()
	internal.MustFPrintf(out,
		`    <circle cx="%d" cy="%d" r="%d" class="%s"></circle>
`,
		pixel.X,
		pixel.Y,
		circleRadius,
		class,
	)
}

func formatMarkBinding(s *markBinding) string {
	return fmt.Sprintf("%+v", s)
}

// Draw a rounded corner as an SVG "elliptical arc" element, here merely a circular arc
// across one of the four axis-aligned quadrants.
//
//   ASCII:
//     Span a _pair_ of left-right adjacent 8x16 cells, one of which contains an ASCII space, that
//     being the one at the corner of the connected line segments.
//            .-.
//           |   |    a circle
//            '-'
//   UTF-8:
//     Corner lies entirely within the _single_ cell containing 'startPixel'
//             ╭╮
//             ╰╯     output will be a true oval -- but no circle is possible
//
func (rc *RoundedCorner) DrawCentered(out io.Writer, centerPixel Pixel, radius int) {
	// https://www.w3.org/TR/SVG/paths.html#PathDataEllipticalArcCommands

	x, y := centerPixel.X, centerPixel.Y
	var startX, startY, sweepFlag, endX, endY int

	switch rc.Orientation {
	case O_NW:
		startX = x
		startY = y - radius
		sweepFlag = 0  // counter-clockwise
		endX = x - radius
		endY = y
	case O_SW:
		startX = x - radius
		startY = y
		sweepFlag = 0  // counter-clockwise
		endX = x
		endY = y + radius
	case O_NE:
		startX = x
		startY = y - radius
		sweepFlag = 1  // clockwise
		endX = x + radius
		endY = y
	case O_SE:
		startX = x + radius
		startY = y
		sweepFlag = 1  // clockwise
		endX = x
		endY = y + radius
	}

	// X  Assumes inherited "fill: none"
	internal.MustFPrintf(out,
		`    <path class="path" d="M %d,%d A %d,%d %d %d,%d %d,%d"></path>
`,
		startX,
		startY,
		radius, // x-radius
		radius, // y-radius
		0, // x-axis-rotation
		0, // large-arc-flag
		sweepFlag,
		endX,  // absolute end position, as implied by SVG command 'A'
		endY,  // absolute end position, as implied by SVG command 'A'
	)
}

// Draw a bridge as an SVG elliptical arc element.
func (b Bridge) Draw(out io.Writer) {
	x, y := b.Start.AsPixelXY()
	sweepFlag := 1

	if b.Orientation == O_W {
		sweepFlag = 0
	}

	// X  Assumes inherited "fill: none"
	internal.MustFPrintf(out,
		`    <path class="path" d="M %d,%d A 9,9 0 0,%d %d,%d"></path>
`,
		x, y-H/2,
		sweepFlag,
		x, y+H/2,
	)
}
