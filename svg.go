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

func (s SVG) String() string {
	return fmt.Sprintf("<svg class='%s' xmlns='%s' version='%s' height='%d' width='%d' font-family='Menlo,Lucida Console,monospace'>\n%s</svg>\n",
		"diagram",
		"http://www.w3.org/2000/svg",
		"1.1", s.Height, s.Width, s.Body)
}

// BuildSVG  reads in a newline-delimited ASCII diagram from src and returns a SVG.
func BuildSVG(src io.Reader) SVG {
	var buff bytes.Buffer
	canvas := NewCanvas(src)
	canvas.WriteSVGBody(&buff)
	return SVG{
		Body:   buff.String(),
		Width:  canvas.widthScreen(),
		Height: canvas.heightScreen(),
	}
}

// BuildAndWriteSVG reads in a newline-delimited ASCII diagram from src and writes a
// corresponding SVG diagram to dst.
func BuildAndWriteSVG(src io.Reader, dst io.Writer) {
	canvas := NewCanvas(src)

	// Preamble
	writeBytes(dst,
		"<svg class='%s' xmlns='%s' version='%s' height='%d' width='%d'>\n",
		"diagram",
		"http://www.w3.org/2000/svg",
		"1.1",
		canvas.heightScreen(), canvas.widthScreen(),
	)

	canvas.WriteSVGBody(dst)

	writeBytes(dst, "</svg>\n")
}

func writeBytes(out io.Writer, format string, args ...interface{}) {
	bytesOut := fmt.Sprintf(format, args...)

	_, err := out.Write([]byte(bytesOut))
	if err != nil {
		panic(nil)
	}
}

// Draw a straight line as an SVG path.
func (l Line) Draw(out io.Writer) {
	start := l.start.asPixel()
	stop := l.stop.asPixel()

	// For cases when a vertical line hits a perpendicular like this:
	//
	//   |          |
	//   |    or    v
	//  ---        ---
	//
	// We need to nudge the vertical line half a vertical cell in the
	// appropriate direction in order to meet up cleanly with the midline of
	// the cell next to it.

	// A diagonal segment all by itself needs to be shifted slightly to line
	// up with _ baselines:
	//     _
	//      \_
	//
	// TODO make this a method on Line to return accurate pixel
	if l.lonely {
		switch l.orientation {
		case NE:
			start.x -= 4
			stop.x -= 4
			start.y += 8
			stop.y += 8
		case SE:
			start.x -= 4
			stop.x -= 4
			start.y -= 8
			stop.y -= 8
		case S:
			start.y -= 8
			stop.y -= 8
		}

		// Half steps
		switch l.chop {
		case N:
			stop.y -= 8
		case S:
			start.y += 8
		}
	}

	if l.needsNudgingDown {
		stop.y += 8
		if l.horizontal() {
			start.y += 8
		}
	}

	if l.needsNudgingLeft {
		start.x -= 8
	}

	if l.needsNudgingRight {
		stop.x += 8
	}

	if l.needsTinyNudgingLeft {
		start.x -= 4
		if l.orientation == NE {
			start.y += 8
		} else if l.orientation == SE {
			start.y -= 8
		}
	}

	if l.needsTinyNudgingRight {
		stop.x += 4
		if l.orientation == NE {
			stop.y -= 8
		} else if l.orientation == SE {
			stop.y += 8
		}
	}

	writeBytes(out,
		"<path d='M %d,%d L %d,%d' fill='none' stroke='currentColor'></path>\n",
		start.x, start.y,
		stop.x, stop.y,
	)
}

// Draw a solid triable as an SVG polygon element.
func (t Triangle) Draw(out io.Writer) {
	// https://www.w3.org/TR/SVG/shapes.html#PolygonElement

	/*
		   	+-----+-----+
		    |    /|\    |
		    |   / | \   |
		  x +- / -+- \ -+
			| /   |   \ |
			|/    |    \|
		    +-----+-----+
		          y
	*/

	x, y := float32(t.start.asPixel().x), float32(t.start.asPixel().y)
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
func (c *Circle) Draw(out io.Writer) {
	fill := "#fff"

	if c.bold {
		fill = "currentColor"
	}

	pixel := c.start.asPixel()

	writeBytes(out,
		"<circle cx='%d' cy='%d' r='6' stroke='currentColor' fill='%s'></circle>\n",
		pixel.x,
		pixel.y,
		fill,
	)
}

// Draw a single text character as an SVG text element.
func (t Text) Draw(out io.Writer) {
	p := t.start.asPixel()
	c := t.contents

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
			p.x-4, p.y-8,
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

	writeBytes(out,
		"<text text-anchor='middle' x='%d' y='%d' fill='currentColor' style='font-size:1em'>%s</text>\n",
		p.x, p.y+4, c,
	)
}

// Draw a rounded corner as an SVG elliptical arc element.
func (c *RoundedCorner) Draw(out io.Writer) {
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
func (b Bridge) Draw(out io.Writer) {
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
