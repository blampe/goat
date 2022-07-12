// All output is buffered into the object SVG, then written to the output stream.
package goat

import (
	"bytes"
	"fmt"
	"io"
)

type SVG struct {
	Body   string
	Width  int
	Height int
}

// See: https://drafts.csswg.org/mediaqueries-5/#prefers-color-scheme
func (s SVG) String(svgColorLightScheme, svgColorDarkScheme string) string {
	style := fmt.Sprintf(
		`<style type="text/css">
svg {
   color: %s;
}
@media (prefers-color-scheme: dark) {
   svg {
      color: %s;
   }
}
</style>`,
		svgColorLightScheme,
		svgColorDarkScheme)

	return fmt.Sprintf(
		"<svg xmlns='%s' version='%s' height='%d' width='%d' font-family='Menlo,Lucida Console,monospace'>\n" +
			"%s\n" +
			"%s</svg>\n",
		"http://www.w3.org/2000/svg",
		"1.1", s.Height, s.Width, style, s.Body)
}

// BuildSVG reads a newline-delimited ASCII diagram from src and returns an
// initialized SVG struct.
func BuildSVG(src io.Reader) SVG {
	var buff bytes.Buffer
	canvas := NewCanvas(src)
	canvas.WriteSVGBody(&buff)
	return SVG{
		Body:	buff.String(),
		Width:	canvas.widthScreen(),
		Height: canvas.heightScreen(),
	}
}

// BuildAndWriteSVG reads in a newline-delimited ASCII diagram from src and writes a
// corresponding SVG diagram to dst.
func BuildAndWriteSVG(src io.Reader, dst io.Writer,
	svgColorLightScheme, svgColorDarkScheme string) {
	svg := BuildSVG(src)
	writeBytes(dst, svg.String(svgColorLightScheme, svgColorDarkScheme))
}

func writeBytes(out io.Writer, format string, args ...interface{}) {
	bytesOut := fmt.Sprintf(format, args...)

	_, err := out.Write([]byte(bytesOut))
	if err != nil {
		panic(err)
	}
}

func writeText(out io.Writer, canvas *Canvas) {
	writeBytes(out,
		`<style>
  text {
       text-anchor: middle;
       font-family: "Menlo","Lucida Console","monospace";
       fill: currentColor;
       font-size: 1em;
  }
</style>
`)
	for _, textObj := range canvas.Text() {
		// usual, baseline case
		textObj.draw(out)
	}
}

// Draw a straight line as an SVG path.
func (l Line) draw(out io.Writer) {
	start := l.start.asPixel()
	stop := l.stop.asPixel()

	// For cases when a vertical line hits a perpendicular like this:
	//
	//   |		|
	//   |	  or	v
	//  ---	       ---
	//
	// We need to nudge the vertical line half a vertical cell in the
	// appropriate direction in order to meet up cleanly with the midline of
	// the cell next to it.

	// A diagonal segment all by itself needs to be shifted slightly to line
	// up with _ baselines:
	//     _
	//	\_
	//
	// TODO make this a method on Line to return accurate pixel
	if l.lonely {
		switch l.orientation {
		case NE:
			start.X -= 4
			stop.X -= 4
			start.Y += 8
			stop.Y += 8
		case SE:
			start.X -= 4
			stop.X -= 4
			start.Y -= 8
			stop.Y -= 8
		case S:
			start.Y -= 8
			stop.Y -= 8
		}

		// Half steps
		switch l.chop {
		case N:
			stop.Y -= 8
		case S:
			start.Y += 8
		}
	}

	if l.needsNudgingDown {
		stop.Y += 8
		if l.horizontal() {
			start.Y += 8
		}
	}

	if l.needsNudgingLeft {
		start.X -= 8
	}

	if l.needsNudgingRight {
		stop.X += 8
	}

	if l.needsTinyNudgingLeft {
		start.X -= 4
		if l.orientation == NE {
			start.Y += 8
		} else if l.orientation == SE {
			start.Y -= 8
		}
	}

	if l.needsTinyNudgingRight {
		stop.X += 4
		if l.orientation == NE {
			stop.Y -= 8
		} else if l.orientation == SE {
			stop.Y += 8
		}
	}

	// If either end is a hollow circle, back off drawing to the edge of the circle,
	// rather extending as usual to center of the cell.
	const (
		ORTHO = 6
		DIAG_X = 3  // XX  By eye, '3' is a bit too much'; '2' is not enough.
		DIAG_Y = 5
	)
	if (l.startRune == 'o') {
		switch l.orientation {
		case NE:
			start.X += DIAG_X
			start.Y -= DIAG_Y
		case E:
			start.X += ORTHO
		case SE:
			start.X += DIAG_X
			start.Y += DIAG_Y
		case S:
			start.Y += ORTHO
		default:
			panic("impossible orientation")
		}
	}
	// X  'stopRune' case differs from 'startRune' only by inversion of the arithmetic signs.
	if (l.stopRune == 'o') {
		switch l.orientation {
		case NE:
			stop.X -= DIAG_X
			stop.Y += DIAG_Y
		case E:
			stop.X -= ORTHO
		case SE:
			stop.X -= DIAG_X
			stop.Y -= DIAG_Y
		case S:
			stop.Y -= ORTHO
		default:
			panic("impossible orientation")
		}
	}

	writeBytes(out,
		"<path d='M %d,%d L %d,%d' fill='none' stroke='currentColor'></path>\n",
		start.X, start.Y,
		stop.X, stop.Y,
	)
}

// Draw a solid triangle as an SVG polygon element.
func (t Triangle) draw(out io.Writer) {
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

	x, y := float32(t.start.asPixel().X), float32(t.start.asPixel().Y)
	r := 0.0

	x0 := x + 8
	y0 := y
	x1 := x - 4
	y1 := y - 0.35*16
	x2 := x - 4
	y2 := y + 0.35*16

	switch t.orientation {
	case N:
		r = 270
		if t.needsNudging {
			x0 += 8
			x1 += 8
			x2 += 8
		}
	case NE:
		r = 300
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case NW:
		r = 240
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case W:
		r = 180
		if t.needsNudging {
			x0 -= 8
			x1 -= 8
			x2 -= 8
		}
	case E:
		r = 0
		if t.needsNudging {
			x0 -= 8
			x1 -= 8
			x2 -= 8
		}
	case S:
		r = 90
		if t.needsNudging {
			x0 += 8
			x1 += 8
			x2 += 8
		}
	case SW:
		r = 120
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case SE:
		r = 60
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	}

	writeBytes(out,
		"<polygon points='%f,%f %f,%f %f,%f' fill='currentColor' transform='rotate(%f, %f, %f)'></polygon>\n",
		x0, y0,
		x1, y1,
		x2, y2,
		r,
		x, y,
	)
}

// Draw a solid circle as an SVG circle element.
func (c *Circle) draw(out io.Writer) {
	var fill string
	if c.bold {
		fill = "currentColor"
	} else {
		fill = "none"
	}
	pixel := c.start.asPixel()
	const circleRadius = 6
	writeBytes(out,
		"<circle cx='%d' cy='%d' r='%d' stroke='currentColor' fill='%s'></circle>\n",
		pixel.X,
		pixel.Y,
		circleRadius,
		fill,
	)
}

// Draw a single text character as an SVG text element.
func (t Text) draw(out io.Writer) {
	p := t.start.asPixel()
	c := t.str

	opacity := 0

	// Markdeep special-cases these character and treats them like a
	// checkerboard.
	switch c {
	case "▉":
		opacity = -64
	case "▓":
		opacity = 64
	case "▒":
		opacity = 128
	case "░":
		opacity = 191
	}

	fill := "currentColor"
	if opacity > 0 {
		fill = fmt.Sprintf("rgb(%d,%d,%d)", opacity, opacity, opacity)
	}

	if opacity != 0 {
		writeBytes(out,
			"<rect x='%d' y='%d' width='8' height='16' fill='%s'></rect>",
			p.X-4, p.Y-8,
			fill,
		)
		return
	}

	// Escape for XML
	switch c {
	case "&":
		c = "&amp;"
	case ">":
		c = "&gt;"
	case "<":
		c = "&lt;"
	}

	// usual case
	writeBytes(out,
		`<text x='%d' y='%d'>%s</text>
`,
		p.X, p.Y+4, c)
}

// Draw a rounded corner as an SVG elliptical arc element.
func (c *RoundedCorner) draw(out io.Writer) {
	// https://www.w3.org/TR/SVG/paths.html#PathDataEllipticalArcCommands

	x, y := c.start.asPixelXY()
	startX, startY, endX, endY, sweepFlag := 0, 0, 0, 0, 0

	switch c.orientation {
	case NW:
		startX = x + 8
		startY = y
		endX = x - 8
		endY = y + 16
	case NE:
		sweepFlag = 1
		startX = x - 8
		startY = y
		endX = x + 8
		endY = y + 16
	case SE:
		sweepFlag = 1
		startX = x + 8
		startY = y - 16
		endX = x - 8
		endY = y
	case SW:
		startX = x - 8
		startY = y - 16
		endX = x + 8
		endY = y
	}

	writeBytes(out,
		"<path d='M %d,%d A 16,16 0 0,%d %d,%d' fill='none' stroke='currentColor'></path>\n",
		startX,
		startY,
		sweepFlag,
		endX,
		endY,
	)
}

// Draw a bridge as an SVG elliptical arc element.
func (b Bridge) draw(out io.Writer) {
	x, y := b.start.asPixelXY()
	sweepFlag := 1

	if b.orientation == W {
		sweepFlag = 0
	}

	writeBytes(out,
		"<path d='M %d,%d A 9,9 0 0,%d %d,%d' fill='none' stroke='currentColor'></path>\n",
		x, y-8,
		sweepFlag,
		x, y+8,
	)
}
