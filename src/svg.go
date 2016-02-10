package goat

import (
	"fmt"
	"io"
)

// ASCIItoSVG reads in a newline-delimited ASCII diagram and writes a
// corresponding SVG diagram.
func ASCIItoSVG(in io.Reader, out io.Writer) {
	canvas := NewCanvas(in)

	// Preamble
	writeBytes(&out,
		"<svg class='%s' xmlns='%s' version='%s' height='%d' width='%d'>\n",
		"diagram",
		"http://www.w3.org/2000/svg",
		"1.1",
		canvas.Height*16+8, (canvas.Width+1)*8,
	)

	writeBytes(&out, "<g transform='translate(8,16)'>\n")

	for _, l := range canvas.Lines() {
		l.Draw(out)
	}

	for _, t := range canvas.Triangles() {
		t.Draw(out)
	}

	for _, c := range canvas.Circles() {
		c.Draw(out)
	}

	for _, c := range canvas.RoundedCorners() {
		c.Draw(out)
	}

	for _, b := range canvas.Bridges() {
		b.Draw(out)
	}

	for _, t := range canvas.Text() {
		t.Draw(out)
	}

	writeBytes(&out, "</g>\n")
	writeBytes(&out, "</svg>\n")
}

func writeBytes(out *io.Writer, format string, args ...interface{}) {
	bytesOut := fmt.Sprintf(format, args...)

	_, err := (*out).Write([]byte(bytesOut))

	if err != nil {
		panic(nil)
	}
}

// Draw a straight line as an SVG path.
func (l *Line) Draw(out io.Writer) {

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

	if l.needsNudgingUp {
		start.y -= 8
	}

	if l.needsNudgingDown {
		stop.y += 8
	}

	writeBytes(&out,
		"<path d='M %d,%d L %d,%d' style='fill:none;stroke:#000;'></path>\n",
		start.x, start.y,
		stop.x, stop.y,
	)
}

// Draw a solid triable as an SVG polygon element.
func (t *Triangle) Draw(out io.Writer) {
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
	case W:
		r = 180
	case S:
		r = 90
		if t.needsNudging {
			x0 += 8
			x1 += 8
			x2 += 8
		}
	}

	writeBytes(&out,
		"<polygon points='%f,%f %f,%f %f,%f' style='fill:#000' transform='rotate(%f, %f, %f)'></polygon>\n",
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
		fill = "#000"
	}

	pixel := c.start.asPixel()

	writeBytes(&out,
		"<circle cx='%d' cy='%d' r='6' style='fill:%s;stroke:#000;'></circle>\n",
		pixel.x,
		pixel.y,
		fill,
	)
}

// Draw a single text character as an SVG text element.
func (t *Text) Draw(out io.Writer) {
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

	if opacity != 0 {
		writeBytes(&out,
			"<rect x='%d' y='%d' width='8' height='16' fill='rgb(%d,%d,%d)'></rect>",
			p.x-4, p.y-8,
			opacity, opacity, opacity,
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

	writeBytes(&out,
		"<text text-anchor='middle' font-family='Menlo,Lucida Console,monospace' x='%d' y='%d' style='fill:#000;font-size:1em'>%s</text>\n",
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

	writeBytes(&out,
		"<path d='M %d,%d A 16,16 0 0,%d %d,%d' style='fill:none;stroke:#000;'></path>\n",
		startX,
		startY,
		sweepFlag,
		endX,
		endY,
	)
}

// Draw a bridge as an SVG elliptical arc element.
func (b *Bridge) Draw(out io.Writer) {
	x, y := b.start.asPixelXY()
	sweepFlag := 1

	if b.orientation == W {
		sweepFlag = 0
	}

	writeBytes(&out,
		"<path d='M %d,%d A 9,9 0 0,%d %d,%d' style='fill:none;stroke:#000;'></path>\n",
		x, y-8,
		sweepFlag,
		x, y+8,
	)
}
