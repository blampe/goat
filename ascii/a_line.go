package ascii

import (
	"io"

	"github.com/blampe/goat"
	"github.com/blampe/goat/svg"
)

// line represents a straight segment between two points 'start' and 'stop', where
// 'start' is either lesser in X (north-east, east, south-east), or
// equal in X and lesser in Y (south).
type line struct {
	// X  These do not necessarily point to 'segment' characters e.g. '-' or '|' -- they
	//    commonly point to adjoining junction ('passThrough') characters e.g. '+'. 
	Start svg.XyIndex
	Stop  svg.XyIndex
	startRune rune
	stopRune rune

	// dashed	    bool
	NeedsNudgingDown      bool   // used for horizontal lines defined with '_'

	// all of these are used to ~extend~ lines so as to complete connection at a corner of a box
	NeedsNudgingLeft      bool   // X  may be combined with NeedsTinyNudgingLeft
	NeedsNudgingRight     bool   // X  may be combined with NeedsTinyNudgingRight
	NeedsTinyNudgingLeft  bool
	NeedsTinyNudgingRight bool

	// This is a line segment all by itself.
	//                        ^^^^^^^^^^^^^   XX What does this mean?    
	//                                             No terminators, at neither end? AXXt one end only?   
	// Centers the segment around the midline.
	Lonely bool

	// N or S. Only useful for half steps - chops off this half of the line.
	Chop svg.Orientation

	// X-major, Y-minor.  Therefore, always one of the compass points NE, E, SE, S.
	Orientation svg.Orientation

	State lineState  // Value is Unstarted or Started
}

const (
	W = svg.CellWidth
	H = svg.CellHeight
)

type lineState int

const (
	unstarted lineState = iota
	started
)

func (l *line) started() bool {
	return l.State == started
}

func (c *Canvas) SetStart(l *line, i svg.XyIndex) {
	if l.State == unstarted {
		l.Start = i
		l.startRune = c.RuneAt(i)
		l.State = started
	}
}

func (c *Canvas) SetStop(l *line, i svg.XyIndex) {
	if l.State == started {
		l.Stop = i
		l.stopRune = c.RuneAt(i)
	}
}

func (l *line) GoesSomewhere() bool {
	return l.Start != l.Stop
}

func (l *line) horizontal() bool {
	return l.Orientation == svg.O_E || l.Orientation == svg.O_W
}

func (l *line) vertical() bool {
	return l.Orientation == svg.O_N || l.Orientation == svg.O_S
}

func (l *line) diagonal() bool {
	return l.Orientation == svg.O_NE || l.Orientation == svg.O_SE || l.Orientation == svg.O_SW || l.Orientation == svg.O_NW
}


// lines returns a slice of all line Drawables,
// in all possible directional orientations, that it can recognize in Canvas.Data[].
func (c *Canvas) lines() (lines []line) {
	horizontalMidlines := c.getlinesForSegment('-')
	diagUplines := c.getlinesForSegment('/')
	for i, l := range diagUplines {
		// /_
		if c.RuneAt(l.Start.East()) == '_' {
			diagUplines[i].NeedsTinyNudgingLeft = true
		}

		// _
		// /
		if c.RuneAt(l.Stop.North()) == '_' {
			diagUplines[i].NeedsTinyNudgingRight = true
		}

		//  _
		// /
		if !l.Lonely && c.RuneAt(l.Stop.NEast()) == '_' {
			diagUplines[i].NeedsTinyNudgingRight = true
		}

		// _/
		if !l.Lonely && c.RuneAt(l.Start.West()) == '_' {
			diagUplines[i].NeedsTinyNudgingLeft = true
		}

		// \
		// /
		if !l.Lonely && c.RuneAt(l.Stop.North()) == '\\' {
			diagUplines[i].NeedsTinyNudgingRight = true
		}

		// /
		// \
		if !l.Lonely && c.RuneAt(l.Start.South()) == '\\' {
			diagUplines[i].NeedsTinyNudgingLeft = true
		}
	}

	diagDownlines := c.getlinesForSegment('\\')
	for i, l := range diagDownlines {
		// _\
		if c.RuneAt(l.Stop.West()) == '_' {
			diagDownlines[i].NeedsTinyNudgingRight = true
		}

		// _
		// \
		if c.RuneAt(l.Start.North()) == '_' {
			diagDownlines[i].NeedsTinyNudgingLeft = true
		}

		//  _
		//   \
		if !l.Lonely && c.RuneAt(l.Start.NWest()) == '_' {
			diagDownlines[i].NeedsTinyNudgingLeft = true
		}

		// \_
		if !l.Lonely && c.RuneAt(l.Stop.East()) == '_' {
			diagDownlines[i].NeedsTinyNudgingRight = true
		}

		// \
		// /
		if !l.Lonely && c.RuneAt(l.Stop.South()) == '/' {
			diagDownlines[i].NeedsTinyNudgingRight = true
		}

		// /
		// \
		if !l.Lonely && c.RuneAt(l.Start.North()) == '/' {
			diagDownlines[i].NeedsTinyNudgingLeft = true
		}
	}

	horizontalBaselines := c.getlinesForSegment('_')
	for i, l := range horizontalBaselines {
		// TODO: make this nudge an orientation
		horizontalBaselines[i].NeedsNudgingDown = true

		//     _
		// _| |      XX  example wrong?  
		if c.RuneAt(l.Stop.SEast()) == '|' || c.RuneAt(l.Stop.NEast()) == '|' {
			horizontalBaselines[i].NeedsNudgingRight = true
		}

		// _
		//  |  _|     XX  example wrong?  
		if c.RuneAt(l.Start.SWest()) == '|' || c.RuneAt(l.Start.NWest()) == '|' {
			horizontalBaselines[i].NeedsNudgingLeft = true
		}

		//     _
		// _/	\     X  example appears right
		if c.RuneAt(l.Stop.East()) == '/' || c.RuneAt(l.Stop.SEast()) == '\\' {
			horizontalBaselines[i].NeedsTinyNudgingRight = true
		}

		//	 _
		// \_	/     X  example appears right
		if c.RuneAt(l.Start.West()) == '\\' || c.RuneAt(l.Start.SWest()) == '/' {
			horizontalBaselines[i].NeedsTinyNudgingLeft = true
		}

		// _\
		if c.RuneAt(l.Stop.East()) == '\\' {
			horizontalBaselines[i].NeedsNudgingRight = true
			horizontalBaselines[i].NeedsTinyNudgingRight = true
		}

		//
		// /_
		if c.RuneAt(l.Start.West()) == '/' {
			horizontalBaselines[i].NeedsNudgingLeft = true
			horizontalBaselines[i].NeedsTinyNudgingLeft = true
		}
		//  _
		//  /
		if c.RuneAt(l.Stop.South()) == '/' {
			horizontalBaselines[i].NeedsTinyNudgingRight = true
		}

		//  _
		//  \
		if c.RuneAt(l.Start.South()) == '\\' {
			horizontalBaselines[i].NeedsTinyNudgingLeft = true
		}

		//  _
		// '
		if c.RuneAt(l.Start.SWest()) == '\'' {
			horizontalBaselines[i].NeedsNudgingLeft = true
		}

		// _
		//  '
		if c.RuneAt(l.Stop.SEast()) == '\'' {
			horizontalBaselines[i].NeedsNudgingRight = true
		}
	}

	verticallines := c.getlinesForSegment('|')

	lines = append(lines, horizontalMidlines...)
	lines = append(lines, horizontalBaselines...)
	lines = append(lines, verticallines...)
	lines = append(lines, diagUplines...)
	lines = append(lines, diagDownlines...)
	lines = append(lines, c.HalfSteps()...)  // vertical, only

	return
}

func newHalfStep(i svg.XyIndex, chop svg.Orientation) line {
	return line{
		Start:	     i,
		Stop:	     i.South(),
		Lonely:	     true,
		Chop:	     chop,
		Orientation: svg.O_S,
	}
}

func (c *Canvas) HalfSteps() (lines []line) {
	for idx := range svg.UpDownMinor(c.Width, c.Height) {
		if o := c.partOfHalfStep(idx); o != svg.O_NONE {
			lines = append(
				lines,
				newHalfStep(idx, o),
			)
		}
	}
	return
}

func (c *Canvas) getlinesForSegment(segment rune) []line {
	var iter svg.CanvasIterator
	var orientation svg.Orientation
	passThroughs := goat.CopySet(jointRunes)

	switch segment {
	case '-':
		iter = svg.LeftRightMinor
		orientation = svg.O_E
		passThroughs.ExtendSet( '<', '>', '(', ')')
	case '_':
		iter = svg.LeftRightMinor
		orientation = svg.O_E
		passThroughs.ExtendSet( '|')
	case '|':   // VERTICAL LINE
		iter = svg.UpDownMinor
		orientation = svg.O_S
		passThroughs.UnionSet(verticalArrowheadRunes)
	case '/':
		iter = svg.DiagUp
		orientation = svg.O_NE
		passThroughs.UnionSet(arrowheadRunes)
		passThroughs.ExtendSet( 'o', '*', '|')
	case '\\':
		iter = svg.DiagDown
		orientation = svg.O_SE
		passThroughs.UnionSet(arrowheadRunes)
		passThroughs.ExtendSet( 'o', '*', '|')
	default:
		return nil
	}

	return c.getlines(segment, iter, orientation, passThroughs)
}

// segment: the primary character expected along a continuing line
// ci: the order that the loop below traverse locations on the canvas.
// o: the orientation for this line.
// passThroughs: characters that will produce a mark that the line segment
//     is allowed to be drawn either through or, in the case of 'o', "underneath" --
//     without terminating the line.
func (c *Canvas) getlines(
	segment rune,
	ci svg.CanvasIterator,
	o svg.Orientation,
	passThroughs goat.RuneSet,
) (lines []line) {
	// Helper to throw the current line we're tracking on to the slice and
	// start a new one.
	snip := func(cl line) line {
		// Only collect lines that actually go somewhere or are isolated
		// segments; otherwise, discard what's been collected so far within 'cl'.
		if cl.GoesSomewhere() {
			lines = append(lines, cl)
		}

		return line{Orientation: o}
	}

	currentline := line{Orientation: o}
	lastSeenRune := ' '

	// X  Purpose of the '+1' overscan is to reset lastSeenRune to ' ' upon wrapping the minor axis.
	for idx := range ci(c.Width+1, c.Height+1) {
		r := c.RuneAt(idx)

		isSegment := r == segment
		isPassThrough := passThroughs.Contains(r)
		isroundedCorner := c.isroundedCorner(idx)
		isDot := isDot(r)
		isTriangle := isTriangle(r)

		justPassedThrough := passThroughs.Contains(lastSeenRune)

		shouldKeep := (isSegment || isPassThrough) && isroundedCorner == svg.O_NONE

		// This is an edge case where we have a rounded corner... that's also a
		// joint... attached to orthogonal line, e.g.:
		//
		//  '+--
		//   |
		//
		// TODO: This also depends on the orientation of the corner and our
		// line.
		// NW / NE line can't go with EW/NS lines, vertical is OK though.
		if isroundedCorner != svg.O_NONE && o != svg.O_E && (c.partOfVerticalline(idx) || c.partOfDiagonalline(idx)) {
			shouldKeep = true
		}

		// Don't connect | to > for diagonal lines or )) for horizontal lines.
		if isPassThrough && justPassedThrough && o != svg.O_S {
			currentline = snip(currentline)
		}

		// Don't connect o to o, + to o, etc. This character is a new pass-through
		// so we still want to respect shouldKeep; we just don't want to draw
		// the existing line through this cell.
		if justPassedThrough && (isDot || isTriangle) {
			currentline = snip(currentline)
		}

		if o == svg.O_S && (r == '.' || lastSeenRune == '\'') {
			currentline = snip(currentline)
		}

		switch currentline.State {
		case unstarted:
			if shouldKeep {
				c.SetStart(&currentline, idx)
				c.SetStop(&currentline, idx)
			}
		case started:
			if !shouldKeep {
				// Snip the existing line, don't add the current cell to it
				// *unless* its a line segment all by itself. If it is, keep a
				// record that it's an individual segment because we need to
				// adjust later in the / and \ cases.
				if !currentline.GoesSomewhere() && lastSeenRune == segment {
					if !c.partOfroundedCorner(currentline.Start) {
						c.SetStop(&currentline, idx)
						currentline.Lonely = true
					}
				}
				currentline = snip(currentline)
			} else if isPassThrough {
				// Snip the existing line but include the current pass-through
				// character because we may be continuing the line.
				c.SetStop(&currentline, idx)
				currentline = snip(currentline)
				c.SetStart(&currentline, idx)
				c.SetStop(&currentline, idx)
			} else if shouldKeep {
				// Keep the line going and extend it by one cell.
				c.SetStop(&currentline, idx)
			}
		}

		lastSeenRune = r
	}
	return
}


// Draw a straight line as an SVG path.
func (l line) Draw(out io.Writer) {
	start := l.Start.AsPixel()
	stop := l.Stop.AsPixel()

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
	// TODO make this a method on line to return accurate pixel
	if l.Lonely {
		switch l.Orientation {
		case svg.O_NE:
			start.X -= W/2
			stop.X -= W/2
			start.Y += H/2
			stop.Y += H/2
		case svg.O_SE:
			start.X -= W/2
			stop.X -= W/2
			start.Y -= H/2
			stop.Y -= H/2
		case svg.O_S:
			start.Y -= H/2
			stop.Y -= H/2
		}

		// Half steps
		switch l.Chop {
		case svg.O_N:
			stop.Y -= H/2
		case svg.O_S:
			start.Y += H/2
		}
	}

	if l.NeedsNudgingDown {
		stop.Y += H/2
		if l.horizontal() {
			start.Y += H/2
		}
	}

	if l.NeedsNudgingLeft {
		start.X -= W
	}

	if l.NeedsNudgingRight {
		stop.X += W
	}

	if l.NeedsTinyNudgingLeft {
		start.X -= W/2
		if l.Orientation == svg.O_NE {
			start.Y += H/2
		} else if l.Orientation == svg.O_SE {
			start.Y -= H/2
		}
	}

	if l.NeedsTinyNudgingRight {
		stop.X += W/2
		if l.Orientation == svg.O_NE {
			stop.Y -= H/2
		} else if l.Orientation == svg.O_SE {
			stop.Y += H/2
		}
	}

	// If either end is a hollow circle, back off drawing to the edge of the circle,
	// rather extending as usual to center of the cell.
	const (
		ORTHO = 6
		DIAG_X = 3  // XX  By eye, '3' is a bit too much'; '2' is not enough.
		DIAG_Y = 5
	)
	if (l.startRune == 'o') {  // XX  ? Easily generalized to needs of BOX chars?   
		switch l.Orientation {
		case svg.O_NE:
			start.X += DIAG_X
			start.Y -= DIAG_Y
		case svg.O_E:
			start.X += ORTHO
		case svg.O_SE:
			start.X += DIAG_X
			start.Y += DIAG_Y
		case svg.O_S:
			start.Y += ORTHO
		default:
			panic("impossible orientation")
		}
	}
	// X  'stopRune' case differs from 'startRune' only by inversion of the arithmetic signs.
	if (l.stopRune == 'o') {  // XX  ? Easily generalized to needs of BOX chars?   
		switch l.Orientation {
		case svg.O_NE:
			stop.X -= DIAG_X
			stop.Y += DIAG_Y
		case svg.O_E:
			stop.X -= ORTHO
		case svg.O_SE:
			stop.X -= DIAG_X
			stop.Y -= DIAG_Y
		case svg.O_S:
			stop.Y -= ORTHO
		default:
			panic("impossible orientation")
		}
	}
	svg.WritePolyline(out, start, stop)
}
