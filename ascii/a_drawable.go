package ascii

import (
	"io"
	//	"log"


	"github.com/blampe/goat/internal"
	"github.com/blampe/goat/svg"
)

// WriteSVGBody writes the entire content of a Canvas out to a stream in SVG format.
// XX Produces a complete <g ...>...</g> expression -- rename accordingly?
func (c *Canvas) WriteSVGBody(dst io.Writer, config *svg.Config) {

	internal.MustFPrintf(dst, "  <g id='%s'>\n", "lines")
	for _, l := range c.lines() {
		l.Draw(dst)
	}

	internal.MustFPrintf(dst, "  </g>\n  <g id='%s'>\n", "triangles")
	for _, tI := range c.triangles() {
		tI.Draw(dst)
	}

	internal.MustFPrintf(dst, "  </g>\n  <g id='%s'>\n", "roundedCorners")
	for _, rc := range c.roundedCorners() {
		const cornerRadius = svg.CellHeight
		drawArc(dst, rc, cornerRadius)
	}

	internal.MustFPrintf(dst, "  </g>\n  <g id='%s'>\n", "circles")
	for _, ci := range c.circles() {
		const circleRadius = 6
		ci.Draw(dst, circleRadius)
	}

	internal.MustFPrintf(dst, "  </g>\n  <g id='%s'>\n", "bridges")
	for _, bI := range c.bridges() {
		bI.Draw(dst)
	}

	internal.MustFPrintf(dst, "  </g>\n  <g id='%s'>\n", "text")
	svg.Writetext(dst, config, c)

	internal.MustFPrintf(dst, "%s", "</g>\n")
}

// triangles detects intended triangles -- typically at the end of an intended line --
// and returns a representational slice composed of the two types Triangle and line.
func (c *Canvas) triangles() (triangles []svg.Drawable) {
	o := svg.O_NONE

	for idx := range svg.UpDownMinor(c.Width, c.Height) {
		needsNudging := false
		start := idx

		r := c.RuneAt(idx)

		if !isTriangle(r) {
			continue
		}

		// Identify orientation and nudge the triangle to touch any
		// adjacent walls.
		switch r {
		case '^':
			o = svg.O_N
			//  ^  and ^
			// /	    \
			if c.RuneAt(start.SWest()) == '/' {
				o = svg.O_NE
			} else if c.RuneAt(start.SEast()) == '\\' {
				o = svg.O_NW
			}
		case 'v':
			if verticalRunes.Contains(c.RuneAt(start.North())) {
				// |
				// v
				o = svg.O_S
			} else if c.RuneAt(start.NEast()) == '/' {
				//  /
				// v
				o = svg.O_SW
			} else if c.RuneAt(start.NWest()) == '\\' {
				//  \
				//   v
				o = svg.O_SE
			} else {
				// Conclusion: Meant as a text string 'v', not a triangle
				//panic("Not sufficient to fix all 'v' troubles.")
				// continue   XX Already committed to non-text output for this string?
				o = svg.O_S
			}
		case '<':
			o = svg.O_W
		case '>':
			o = svg.O_E
		}

		// Determine if we need to snap the triangle to something and, if so,
		// draw a tail if we need to.
		switch o {
		case svg.O_N:
			r := c.RuneAt(start.North())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(triangles, newHalfStep(start, svg.O_N))
			}
		case svg.O_NW:
			r := c.RuneAt(start.NWest())
			// Need to draw a tail.
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					line{
						Start:	     start.NWest(),
						Stop:	     start,
						Orientation: svg.O_SE,
					},
				)
			}
		case svg.O_NE:
			r := c.RuneAt(start.NEast())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					line{
						Start:	     start,
						Stop:	     start.NEast(),
						Orientation: svg.O_NE,
					},
				)
			}
		case svg.O_S:
			r := c.RuneAt(start.South())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(triangles, newHalfStep(start, svg.O_S))
			}
		case svg.O_SE:
			r := c.RuneAt(start.SEast())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					line{
						Start:	     start,
						Stop:	     start.SEast(),
						Orientation: svg.O_SE,
					},
				)
			}
		case svg.O_SW:
			r := c.RuneAt(start.SWest())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					line{
						Start:	     start.SWest(),
						Stop:	     start,
						Orientation: svg.O_NE,
					},
				)
			}
		case svg.O_W:
			r := c.RuneAt(start.West())
			if isDot(r) {
				needsNudging = true
			}
		case svg.O_E:
			r := c.RuneAt(start.East())
			if isDot(r) {
				needsNudging = true
			}
		}

		triangles = append(
			triangles,
			svg.Triangle{
				Start:	      start,
				Orientation:  o,
				NeedsNudging: needsNudging,
			},
		)
	}
	return
}

// circles returns a slice of all 'o' and '*' characters not considered text.
func (c *Canvas) circles() (circles []svg.Circle) {
	for idx := range svg.UpDownMinor(c.Width, c.Height) {
		// TODO INCOMING
		if c.RuneAt(idx) == 'o' {
			circles = append(circles, svg.Circle{Start: idx})
		} else if c.RuneAt(idx) == '*' {
			circles = append(circles, svg.Circle{Start: idx, Bold: true})
		}
	}
	return
}

// roundedCorners returns a slice of all curvy corners in the diagram.
func (c *Canvas) roundedCorners() (corners []svg.RoundedCorner) {
	for idx := range svg.LeftRightMinor(c.Width, c.Height) {
		if o := c.isroundedCorner(idx); o != svg.O_NONE {
			corners = append(
				corners,
				svg.RoundedCorner{Start: idx, Orientation: o},
			)
		}
	}
	return
}

// For . and ' characters this will return a non-svg.O_NONE orientation if the
// contents of adjacent characters satisfy the rules to allow a rounded corner,
// in particular that aa circular arc can be drawn to connect a vertical edge
// with a horizontal edge.
func (c *Canvas) isroundedCorner(i svg.XyIndex) svg.Orientation {
	r := c.RuneAt(i)

	if !isJoint(r) {
		return svg.O_NONE
	}

	left := i.West()
	right := i.East()
	lowerLeft := i.SWest()
	lowerRight := i.SEast()
	upperLeft := i.NWest()
	upperRight := i.NEast()

	opensUp := r == '\'' || r == '+'
	opensDown := r == '.' || r == '+'

	dashRight := c.RuneAt(right) == '-' || c.RuneAt(right) == '+' || c.RuneAt(right) == '_' || c.RuneAt(upperRight) == '_'
	dashLeft := c.RuneAt(left) == '-' || c.RuneAt(left) == '+' || c.RuneAt(left) == '_' || c.RuneAt(upperLeft) == '_'

	isVerticalSegment := func(i svg.XyIndex) bool {
		r := c.RuneAt(i)
		return verticalRunes.Contains(r) || r == '+' || r == ')' || r == '(' || isDot(r)
	}

	//  .- or  .-
	// |	  +
	if opensDown && dashRight && isVerticalSegment(lowerLeft) {
		return svg.O_NW
	}

	// -. or -.  or -.  or _.  or -.
	//   |	   +	  )	 )	o
	if opensDown && dashLeft && isVerticalSegment(lowerRight) {
		return svg.O_NE
	}

	//   | or   + or   | or	  + or	 + or_ )
	// -'	  -'	 +'	+'     ++     '
	if opensUp && dashLeft && isVerticalSegment(upperRight) {
		return svg.O_SE
	}

	// |  or +
	//  '-	  '-
	if opensUp && dashRight && isVerticalSegment(upperLeft) {
		return svg.O_SW
	}

	return svg.O_NONE
}

// bridges returns a slice of all bridges, "-)-" or "-(-", composed as a sequence of
// either type bridge or type line.
func (c *Canvas) bridges() (bridges []svg.Drawable) {
	for idx := range svg.LeftRightMinor(c.Width, c.Height) {
		if o := c.isBridge(idx); o != svg.O_NONE {
			bridges = append(
				bridges,
				newHalfStep(idx.North(), svg.O_S),
				newHalfStep(idx.South(), svg.O_N),
				svg.Bridge{
					Start:	     idx,
					Orientation: o,
				},
			)
		}
	}
	return
}

// -)- or -(- or
func (c *Canvas) isBridge(i svg.XyIndex) svg.Orientation {
	r := c.RuneAt(i)

	left := c.RuneAt(i.West())
	right := c.RuneAt(i.East())

	if left != '-' || right != '-' {
		return svg.O_NONE
	}

	if r == '(' {
		return svg.O_W
	}

	if r == ')' {
		return svg.O_E
	}

	return svg.O_NONE
}


