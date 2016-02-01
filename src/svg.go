package goaat

import (
	"fmt"
	"io"
)

func ASCIItoSVG(in io.Reader, out io.Writer) {
	canvas := NewCanvas(in)

	out.Write(
		[]byte(fmt.Sprintf(
			"<svg class='diagram' xmlns='http://www.w3.org/2000/svg' version='1.1' height='%d' width='%d'>\n",
			canvas.Height*16, canvas.Width*8,
		)),
	)

	out.Write([]byte("<g transform='translate(8,16)'>\n"))

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

	out.Write([]byte("</g>\n"))
	out.Write([]byte("</svg>\n"))
}

func (l *Line) Draw(out io.Writer) {

	start := l.start.asPixel()
	stop := l.stop.asPixel()

	out.Write([]byte(fmt.Sprintf(
		"<path d='M %d,%d L %d,%d' style='fill:none;stroke:#000;'></path>\n",
		start.x, start.y,
		stop.x, stop.y,
	)))
}

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
		//x0 += 8
		//x1 += 8
		//x2 += 8
	case W:
		r = 180
	case S:
		r = 90
		//x0 += 8
		//x1 += 8
		//x2 += 8
	}

	out.Write([]byte(fmt.Sprintf(
		"<polygon points='%f,%f %f,%f %f,%f' style='fill:#000' transform='rotate(%f, %f, %f)'></polygon>\n",
		x0, y0,
		x1, y1,
		x2, y2,
		r,
		x, y,
	)))
}

func (c *Circle) Draw(out io.Writer) {
	fill := "#fff"

	if c.bold {
		fill = "#000"
	}

	pixel := c.start.asPixel()

	out.Write([]byte(fmt.Sprintf(
		"<circle cx='%d' cy='%d' r='6' style='fill:%s;stroke:#000;'></circle>\n",
		pixel.x,
		pixel.y,
		fill,
	)))
}

func (t *Text) Draw(out io.Writer) {
	p := t.start.asPixel()
	char := t.contents

	switch char {
	case "&":
		char = "&amp;"
	case ">":
		char = "&gt;"
	case "<":
		char = "&lt;"
	}

	out.Write([]byte(fmt.Sprintf(
		"<text text-anchor='middle' x='%d' y='%d' style='fill:#000'>%s</text>\n",
		p.x, p.y, char,
	)))
}

func (c *RoundedCorner) Draw(out io.Writer) {
	// https://www.w3.org/TR/SVG/paths.html#PathDataEllipticalArcCommands

	x, y := c.start.asPixel().x, c.start.asPixel().y
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
	out.Write([]byte(fmt.Sprintf(
		"<path d='M %d,%d A 16,16 0 0,%d %d,%d' style='fill:none;stroke:#000;'></path>\n",
		startX,
		startY,
		sweepFlag,
		endX,
		endY,
	)))
}

func (b *Bridge) Draw(out io.Writer) {
}
