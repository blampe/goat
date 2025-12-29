package utf8

import (
	"io"

	"github.com/blampe/goat/internal"
	"github.com/blampe/goat/svg"
)

const (
	W = svg.CellWidth   // XX DRY with hard-coded constants in a_line.go et al.   
	H = svg.CellHeight

	CIRCLERADIUS = W/2
	cornerRadius = W/2
)

// WriteSVGBody writes the entire content of a Canvas out to a stream in SVG format.
// XX Produces a complete <g ...>...</g> expression -- rename accordingly?
func (c *Canvas) WriteSVGBody(dst io.Writer, config *svg.Config) {
	// XX  Refactor for readability.
	wb := func(fmt string, s ...interface{}) {
		internal.MustFPrintf(dst, fmt, s...)
	}

	wb("  <g id='%s'>\n", "lines-vertical")
	for _, lv := range c.getlines(svg.UpDownMinor, svg.O_S) {
		c.DrawLine(lv, dst)
	}
	wb("  </g>\n")

	wb("  <g id='%s'>\n", "lines-horizontal")
	for _, lh := range c.getlines(svg.LeftRightMinor, svg.O_E) {
		c.DrawLine(lh, dst)
	}
	wb("  </g>\n")

	// XX unify with '/ascii'
	wb("  <g id='%s'>\n", "triangles")
	for _, t := range c.triangles() {
		t.Draw(dst)
	}
	wb("  </g>\n")

	// Unicode's tightly-rounded "BOX LIGHT" corners, as
	// parallel to Ascii-mode's widely-rounded.
	// XX  Not as easily discoverable as wanted:
	//      An svg.RoundedCorner struct contains two fields, Start and Orientation
	wb("  <g id='%s'>\n", "roundedCorners")
	for _, rc := range c.roundedCorners() {
		rc.DrawCentered(dst, c.CenterPixel(rc), cornerRadius)
	}
	wb("  </g>\n")

	// XX unify with '/ascii'
	wb("  <g id='%s'>\n", "circles")
	for _, ci := range c.circles() {
		ci.Draw(dst, CIRCLERADIUS)
	}
	wb("  </g>\n")

	wb("  <g id='%s'>\n", "text")
	svg.Writetext(dst, config, c)
	wb("  </g>\n")
}

func (c *Canvas) triangles() (triangles []smallTriangle) {
	for idx := range svg.UpDownMinor(c.Width, c.Height) {
		r := c.RuneAt(idx)

		o := svg.O_NONE
		// Identify orientation and nudge the triangle to touch any
		// adjacent walls.
		switch r {
		case '▲':
			o = svg.O_N
		case '▼':
			o = svg.O_S
		case '◀':
			fallthrough
		case '◄':
			o = svg.O_W
		case '▶':
			fallthrough
		case '►':
			o = svg.O_E
		default:
			continue
		}
		start := idx
		triangles = append(
			triangles,
			smallTriangle{
				Start:	      start,
				Orientation:  o,
				//NeedsNudging: needsNudging,
			},
		)
	}
	return
}

func (c *Canvas) circles() (circles []svg.Circle) {
	for idx := range svg.UpDownMinor(c.Width, c.Height) {
		// TODO INCOMING
		if c.RuneAt(idx) == '○' {
			circles = append(circles, svg.Circle{Start: idx})
		} else if c.RuneAt(idx) == '●' {
			circles = append(circles, svg.Circle{Start: idx, Bold: true})
		}
	}
	return
}

// XX  Exact copy of ascii.roundedCorners(), except that Canvas are different types => DRY?
// roundedCorners returns a slice of all curvy corners in the diagram.
func (c *Canvas) roundedCorners() (corners []svg.RoundedCorner) {
	for idx := range svg.LeftRightMinor(c.Width, c.Height) {
		if o := c.isroundedCorner(idx); o != svg.O_NONE {  // XX specific to 'utf8'
			// XX  demultiplex into an array by Orientation of slices, right here?
			corners = append(
				corners,
				svg.RoundedCorner{Start: idx, Orientation: o},
			)
		}
	}
	return
}

// cf: 'var roundedCornerRunes'
func (c *Canvas) isroundedCorner(i svg.XyIndex) svg.Orientation {
	r := c.RuneAt(i)
	switch r {
		case '╭': return svg.O_NW
		case '╰': return svg.O_SW
		case '╮': return svg.O_NE
		case '╯': return svg.O_SE
	}
	return svg.O_NONE
}


