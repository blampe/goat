package utf8

import (
	"io"

	"github.com/blampe/goat/svg"
)

// Principle: Treat the two axes completely independently.
// Therefore, input "┼─┬" produces three line structs.
//
type line struct {  // XX  take local
	Started bool  // XX  could be local variable of constructor loop?

	// These point to 'segment' characters only in the case of "bare end" --
	// more commonly they point to adjoining junction characters.
 	Start, Stop svg.XyIndex

	// Always one of the compass points O_E, O_S.
	Orientation svg.Orientation
}

func (c *Canvas) SetStart(l *line, i svg.XyIndex) {
	if l.Started {
		panic("Already started")
	}
	l.Start = i
	l.Started = true
}

func (c *Canvas) SetStop(l *line, i svg.XyIndex) {
	if ! l.Started {
		panic("not started")
	}
	l.Stop = i
}

func reverse(o svg.Orientation) svg.Orientation {
	switch o {
	case svg.O_E:
		return svg.O_W
	case svg.O_S:
		return svg.O_N
	}
	panic("unexpected svg.Orientation")
}

// Recall that spatial precision of output line{} is whole character cells.
//
// In the baseline case, an SVG line is to be drawn from the center of line.Start to center of line.End .
// Adjustments to abut well with neighbors happen later.
func (c *Canvas) getlines(
	ci svg.CanvasIterator,  // the order that the loop below traverses cells on the canvas.
	minor_axis svg.Orientation,  // either O_E or O_S
) (lines []line) {

	reverse := reverse(minor_axis) // either O_W or O_N

	// Write 'currentLine' onto the output.
	// line may be of apparently zero-length i.e. contained within a single cell.
	outputLine := func(l line) line {
			lines = append(lines, l)

			// start a new one.
			// XX  Writing the orientation into the line struct is
			//     redundant (now, with separated slices).
			return line{Orientation: minor_axis}
	}

	currentline := line{Orientation: minor_axis}
	for idx := range ci(c.Width+1, c.Height+1) {
		r := c.RuneAt(idx)

		if !currentline.Started {
			if connects[reverse].Contains(r) {
				// Half-cell-long segment
				c.SetStart(&currentline, idx)
				c.SetStop(&currentline, idx)
				continue
			}
			if connects[minor_axis].Contains(r) {
				c.SetStart(&currentline, idx)
				c.SetStop(&currentline, idx)
			}
			continue
		}

		if connects[minor_axis].Contains(r) {
			// Keep the line going and extend it by one cell.
			c.SetStop(&currentline, idx)
		} else {
			if connects[reverse].Contains(r) {
				// Terminate the line at a BOX T-intersection, or a triangle
				c.SetStop(&currentline, idx)
			} else {
				// terminate at a dead end
				// Stop is as set in previous iteration
			}
			currentline = outputLine(currentline)
		}
	}
	return
}

// Draw a straight line as an SVG path.
//
// Cases:
//   1. Unterminated lines composed of '─' or '│' faithfully -- extend to
//      limit of end cells, just as for interior '─' or '│'.
//
//   2. If a terminal circle or square, possibly open, lies at either end,
//      extend so as to abut there.
//
//   3. Isolated joints e.g. '┬' or '┼': Decompose into horizontal and vertical,
//      each independent of the other.

func (c *Canvas) DrawLine(l line, out io.Writer) {
	startPix := c.startingPixel(l)
	stopPix := c.stoppingPixel(l)

	if startPix.X == stopPix.X && startPix.Y == stopPix.Y {
		return
	}

	svg.WritePolyline(out, startPix, stopPix)
}

func (c *Canvas) startingPixel(l line) svg.Pixel {
	// initial values of these are at centers of cells -- possibly adjusted later
	startPix := l.Start.AsPixel()
	startRune := c.RuneAt(l.Start)

	switch l.Orientation {
	case svg.O_E:
		if startRune == '╭' || startRune == '╰' {
			startPix.X += cornerRadius
		} else if connects[reverse(l.Orientation)].Contains(startRune) {
			westRune := c.RuneAt(l.Start.West())
			startTriangleBase := leftArrowheadRunes.Contains(westRune)

			startPix.X -= W/2
			if startTriangleBase {
				startPix.X -= W/4
			}
		}
	case svg.O_S:
		if startRune == '╭' || startRune == '╮' {
			startPix.Y += cornerRadius
		} else if connects[reverse(l.Orientation)].Contains(startRune) {
			// If either end abuts a circle, extend drawing to the edge of the circle,
			// rather extending as usual to center of the cell.
			northRune := c.RuneAt(l.Start.North())
			startTriangleBase := northRune == '▲'

			startPix.Y -= H/2
			if isDot(northRune) {
				startPix.Y -= H/4
			} else if startTriangleBase {
				startPix.Y -= H/1
			}
		}
	}
	return startPix
}

func (c *Canvas) stoppingPixel(l line) svg.Pixel {
	// initial values of these are at centers of cells -- possibly adjusted later
	stopPix := l.Stop.AsPixel()
	stopRune := c.RuneAt(l.Stop)

	switch l.Orientation {
	case svg.O_E:
		if stopRune == '╮' || stopRune == '╯' {
			stopPix.X -= cornerRadius
		} else if connects[l.Orientation].Contains(stopRune) {
			eastRune := c.RuneAt(l.Stop.East())
			stopTriangleBase := rightArrowheadRunes.Contains(eastRune)

			// extend to edge of cell
			stopPix.X += W/2
			if stopTriangleBase {
				stopPix.X += W/4
			}
		}
	case svg.O_S:
		if stopRune == '╯' || stopRune == '╰' {
			stopPix.Y -= cornerRadius
		} else if connects[l.Orientation].Contains(stopRune) {
			// If either end abuts a circle, extend drawing to the edge of the circle,
			// rather extending as usual to center of the cell.
			southRune := c.RuneAt(l.Stop.South())
			stopTriangleBase := southRune == '▼'

			// draw from center to "forward" edge of cell
			stopPix.Y += H/2
			if isDot(southRune) {
				stopPix.Y += H/4
			} else if stopTriangleBase {
				stopPix.Y += H/1
			}
		}
	}
	return stopPix
}
