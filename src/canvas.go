package goat

import (
	"bufio"
	"bytes"
	"io"
)

var jointRunes = []rune{'.', '\'', '+'}
var reservedRunes = map[rune]bool{
	'-':  true,
	'|':  true,
	'v':  true,
	'^':  true,
	'>':  true,
	'<':  true,
	'o':  true,
	'*':  true,
	'+':  true,
	'.':  true,
	'\'': true,
	'/':  true,
	'\\': true,
	')':  true,
	'(':  true,
	'╱':  true,
	'╲':  true,
	'╳':  true,
}

func contains(in []rune, r rune) bool {
	for _, v := range in {
		if r == v {
			return true
		}
	}
	return false
}

// Canvas represents a 2D ASCII rectangle.
type Canvas struct {
	Width  int
	Height int
	data   map[Index]rune
}

func (c *Canvas) String() string {
	var buffer bytes.Buffer

	for h := 0; h < c.Height; h++ {
		for w := 0; w < c.Width; w++ {
			idx := Index{w, h}
			_, err := buffer.WriteRune(c.runeAt(idx))
			if err != nil {
				continue
			}
		}

		err := buffer.WriteByte('\n')
		if err != nil {
			continue
		}
	}

	return buffer.String()
}

func (c *Canvas) runeAt(i Index) rune {

	if val, ok := c.data[i]; ok {
		return val
	}

	return ' '
}

// NewCanvas creates a new canvas with contents read from the given io.Reader.
// Content should be newline delimited.
func NewCanvas(in io.Reader) Canvas {
	width := 0
	height := 0

	scanner := bufio.NewScanner(in)

	data := make(map[Index]rune)

	for scanner.Scan() {
		line := scanner.Text()

		w := 0
		// Can't use index here because it corresponds to unicode offsets
		// instead of logical characters.
		for _, c := range line {
			idx := Index{x: w, y: height}
			data[idx] = rune(c)
			w++
		}

		if w > width {
			width = w
		}
		height++
	}

	return Canvas{Width: width, Height: height, data: data}
}

// Drawable represents anything that can Draw itself.
type Drawable interface {
	Draw(out io.Writer)
}

// Line represents a straight segment between two points.
type Line struct {
	start Index
	stop  Index
	//dashed           bool
	needsNudgingUp   bool
	needsNudgingDown bool

	state lineState
}

type lineState int

const (
	_Unstarted lineState = iota
	_Started
)

func (l *Line) started() bool {
	return l.state == _Started
}

func (l *Line) setStart(i Index) {
	if l.state == _Unstarted {
		l.start = i
		l.stop = i
		l.state = _Started
	}
}

func (l *Line) setStop(i Index) {
	if l.state == _Started {
		l.stop = i
	}
}

// Triangle corresponds to "^", "v", "<" and ">" runes in the absence of
// surrounding alphanumerics.
type Triangle struct {
	start        Index
	orientation  Orientation
	needsNudging bool
}

// Circle corresponds to "o" or "*" runes in the absence of surrounding
// alphanumerics.
type Circle struct {
	start Index
	bold  bool
}

// RoundedCorner corresponds to combinations of "-." or "-'".
type RoundedCorner struct {
	start       Index
	orientation Orientation
}

// Text corresponds to any runes not reserved for diagrams, or reserved runes
// surrounded by alphanumerics.
type Text struct {
	start    Index
	contents string
}

// Bridge correspondes to combinations of "-)-" or "-(-" and is displayed as
// the vertical line "hopping over" the horizontal.
type Bridge struct {
	start       Index
	orientation Orientation
}

// Orientation represents the primary direction that a Drawable is facing.
type Orientation int

const (
	NONE Orientation = iota // No orientation; no structure present.
	N                       // North
	NE                      // Northeast
	NW                      // Northwest
	S                       // South
	SE                      // Southeast
	SW                      // Southwest
	E                       // East
	W                       // West
)

// Lines returns a slice of all Line drawables that we can detect -- in all
// possible orientations.
func (c *Canvas) Lines() []Line {

	lines := c.linesFromIterator(
		upDown,
		[]rune{'|'},
		append([]rune{'v', '^', 'o', '*'}, jointRunes...),
	)

	for i, l := range lines {
		above := c.runeAt(l.start.north())
		below := c.runeAt(l.stop.south())
		if (c.runeAt(l.start) == '|' && above == '-' || above == '(' || above == ')') || c.runeAt(l.start) == '^' {
			lines[i].needsNudgingUp = true
		}
		if (c.runeAt(l.stop) == '|' && below == '-' || below == ')' || below == '(') || c.runeAt(l.stop) == 'v' {
			lines[i].needsNudgingDown = true
		}
	}

	lines = append(lines, c.linesFromIterator(
		leftRight,
		[]rune{'-', ')', '('},
		append([]rune{'o', '*', '<', '>'}, jointRunes...),
	)...)

	lines = append(lines, c.linesFromIterator(
		diagUp,
		[]rune{'/', '╱', '╳'},
		append([]rune{'o', '*', '<', '>', '^', 'v', '|'}, jointRunes...),
	)...)

	lines = append(lines, c.linesFromIterator(
		diagDown,
		[]rune{'\\', '╲', '╳'},
		append([]rune{'o', '*', '<', '>', '^', 'v', '|'}, jointRunes...),
	)...)

	return lines
}

// ci: the order that we traverse locations on the canvas.
// segmentPieces characters we 1) include, and 2) keep going.
// inclusiveTerminals: characters we 1) include, and 2) end the current line.
// exclusiveTerminals: characters we 1) don't include, and 2) end the line.
func (c *Canvas) linesFromIterator(
	ci canvasIterator,
	segments []rune,
	terminals []rune,
) []Line {
	var lines []Line

	var currentLine Line
	var lastSeenRune rune

	// Helper to throw the current line we're tracking on to the slice and
	// start a new one.
	snip := func(l Line) Line {
		lines = append(lines, l)
		return Line{}
	}

	for idx := range ci(c.Width, c.Height) {
		r := c.runeAt(idx)

		isText := c.isText(idx)
		isTerminal := contains(terminals, r)
		isSegment := contains(segments, r)
		isRoundedCorner := c.isRoundedCorner(idx) != NONE
		isDot := r == 'o' || r == '*'
		isTriangle := r == '^' || r == 'v' || r == '<' || r == '>'

		justSawATerminal := contains(terminals, lastSeenRune)

		shouldKeep := (isSegment || isTerminal) && !isText && !isRoundedCorner

		// Don't connect | to > for diagonal lines.
		if isTerminal && justSawATerminal && !contains(segments, '|') {
			currentLine = snip(currentLine)
		}

		// Don't connect o to o, + to o, etc. This character is a new terminal
		// so we still want to respect shouldKeep; we just don't want to draw
		// the existing line through this cell.
		if justSawATerminal && (isDot || isTriangle) {
			currentLine = snip(currentLine)
		}

		switch currentLine.state {
		case _Unstarted:
			if shouldKeep {
				currentLine.setStart(idx)
			}
		case _Started:
			if !shouldKeep {
				// Snip the existing line, don't add the current cell to it.
				currentLine = snip(currentLine)
			} else if isTerminal {
				// Snip the existing line but include the current terminal
				// cell.
				currentLine.setStop(idx)
				currentLine = snip(currentLine)
				currentLine.setStart(idx)
			} else if shouldKeep {
				// Keep the line going and extend it by this character.
				currentLine.setStop(idx)
			}
		}

		lastSeenRune = r
	}

	return lines
}

// Triangles returns a slice of all detectable Triangles.
func (c *Canvas) Triangles() []Triangle {
	var triangles []Triangle

	o := NONE

	for idx := range upDown(c.Width, c.Height) {
		needsNudging := false
		start := idx

		if c.isText(idx) {
			continue
		}

		r := c.runeAt(idx)

		// Identify our orientation and nudge the triangle to touch any
		// adjacent walls.
		switch r {
		case '^':
			o = N
			r := c.runeAt(start.north())
			if r == '-' || contains(jointRunes, r) {
				needsNudging = true
			}
		case 'v':
			o = S
			r := c.runeAt(start.south())
			if r == '-' || contains(jointRunes, r) {
				needsNudging = true
			}
		case '<':
			o = W
		case '>':
			o = E
		default:
			continue
		}

		triangles = append(
			triangles,
			Triangle{start: start, orientation: o, needsNudging: needsNudging},
		)
	}

	return triangles
}

// Circles returns a slice of all 'o' and '*' characters not considered text.
func (c *Canvas) Circles() []Circle {
	var circles []Circle

	for idx := range upDown(c.Width, c.Height) {
		// TODO INCOMING
		if c.runeAt(idx) == 'o' && !c.isText(idx) {
			circles = append(circles, Circle{start: idx})
		} else if c.runeAt(idx) == '*' && !c.isText(idx) {
			circles = append(circles, Circle{start: idx, bold: true})
		}
	}

	return circles
}

// RoundedCorners returns a slice of all curvy corners in the diagram.
func (c *Canvas) RoundedCorners() []RoundedCorner {
	var corners []RoundedCorner

	for idx := range leftRight(c.Width, c.Height) {
		if o := c.isRoundedCorner(idx); o != NONE {
			corners = append(
				corners,
				RoundedCorner{start: idx, orientation: o},
			)
		}
	}

	return corners
}

// For . and ' characters this will return a non-NONE orientation if the
// character falls on a rounded corner.
func (c *Canvas) isRoundedCorner(i Index) Orientation {
	r := c.runeAt(i)

	if r != '.' && r != '\'' {
		return NONE
	}

	left := i.west()
	right := i.east()
	lowerLeft := i.sWest()
	lowerRight := i.sEast()
	upperLeft := i.nWest()
	upperRight := i.nEast()

	dashRight := c.runeAt(right) == '-' || c.runeAt(right) == '+'
	dashLeft := c.runeAt(left) == '-' || c.runeAt(left) == '+'

	isVerticalSegment := func(i Index) bool {
		r := c.runeAt(i)
		return r == '|' || r == '+'
	}

	if r == '.' {
		// North case

		//  .- or  .-
		// |      +
		if dashRight && isVerticalSegment(lowerLeft) {
			return NW
		}

		// -. or -.
		//   |     +
		if dashLeft && isVerticalSegment(lowerRight) {
			return NE
		}

	} else {
		// South case

		//   | or   +
		// -'     -'
		if dashLeft && isVerticalSegment(upperRight) {
			return SE
		}

		// |  or +
		//  '-    '-
		if dashRight && isVerticalSegment(upperLeft) {
			return SW
		}
	}

	return NONE
}

// Text returns a slace of all text characters not belonging to part of the diagram.
// How these characters are identified is rather complicated.
func (c *Canvas) Text() []Text {
	var text []Text

	for i := range leftRight(c.Width, c.Height) {

		if c.isText(i) {
			r := c.runeAt(i)
			text = append(text, Text{start: i, contents: string(r)})
		}

	}
	return text
}

// Bridges returns a slice of all bridges, "-)-" or "-(-".
func (c *Canvas) Bridges() []Bridge {
	var bridges []Bridge

	for idx := range leftRight(c.Width, c.Height) {
		if o := c.isBridge(idx); o != NONE {
			bridges = append(bridges, Bridge{start: idx, orientation: o})
		}
	}

	return bridges
}

// -)- or -(- or
func (c *Canvas) isBridge(i Index) Orientation {

	r := c.runeAt(i)

	left := c.runeAt(i.west())
	right := c.runeAt(i.east())

	if left != '-' || right != '-' {
		return NONE
	}

	if r == '(' {
		return W
	}

	if r == ')' {
		return E
	}

	return NONE
}

func (c *Canvas) isText(i Index) bool {

	if !c.withinBounds(i) {
		return false
	}

	// This index refers to a rune not in our reserved set.
	if c.isDefinitelyText(i) {
		return true
	}

	// This is a reserved character with an incoming line (e.g., "|") above it,
	// so call it non-text.
	if c.hasLineAboveOrBelow(i) {
		return false
	}

	// Reserved characters like "o" or "*" with letters sitting next to them
	// are probably text.
	if c.isTextLeft(i, 2) || c.isTextRight(i, 2) {
		return true
	}

	return false
}

func (c *Canvas) isTextLeft(i Index, limit uint8) bool {
	if limit == 0 {
		return false
	}
	left := i.west()

	return c.isDefinitelyText(left) || c.isTextLeft(left, limit-1)
}

func (c *Canvas) isTextRight(i Index, limit uint8) bool {
	if limit == 0 {
		return false
	}
	right := i.east()

	return c.isDefinitelyText(right) || c.isTextRight(right, limit-1)
}

// Returns true if the character at this index is not reserved for diagrams.
// Characters like "o" need more context (e.g., are other text characters
// nearby) to determine whether they're part of a diagram.
func (c *Canvas) isDefinitelyText(i Index) bool {
	r := c.runeAt(i)

	if r == ' ' {
		return false
	}

	_, isReserved := reservedRunes[r]

	return !isReserved
}

func (c *Canvas) hasLineAboveOrBelow(i Index) bool {
	r := c.runeAt(i)

	nEast := i.nEast()
	sWest := i.sWest()

	switch r {
	case '*', 'o', '+':
		return c.partOfDiagonalLine(i) || c.partOfVerticalLine(i)
	case '|':
		return c.partOfVerticalLine(i) || c.partOfRoundedCorner(i)
	case '/':
		return c.partOfDiagonalLine(i) || contains(jointRunes, c.runeAt(nEast)) || contains(jointRunes, c.runeAt(sWest))
	case '-':
		return c.partOfRoundedCorner(i)
	case '(', ')':
		return c.partOfVerticalLine(i)
	}

	return false
}

// Returns true if a "|" segment passes through this index.
func (c *Canvas) partOfVerticalLine(i Index) bool {
	this := c.runeAt(i)
	north := c.runeAt(i.north())
	south := c.runeAt(i.south())

	jointAboveMe := this == '|' && contains(jointRunes, north)

	if north == '|' || jointAboveMe {
		return true
	}

	jointBelowMe := this == '|' && contains(jointRunes, south)

	if south == '|' || jointBelowMe {
		return true
	}

	return false
}

// Return true if a "--" segment passes through this index.
func (c *Canvas) partOfHorizontalLine(i Index) bool {
	return c.runeAt(i.east()) == '-' || c.runeAt(i.west()) == '-'
}

func (c *Canvas) partOfDiagonalLine(i Index) bool {
	return (c.runeAt(i.nWest()) == '\\' ||
		c.runeAt(i.sEast()) == '\\' ||
		c.runeAt(i.nEast()) == '/' ||
		c.runeAt(i.sWest()) == '/')
}

// For "-" and "|" characters returns true if they could be part of a rounded
// corner.
func (c *Canvas) partOfRoundedCorner(i Index) bool {
	r := c.runeAt(i)

	switch r {
	case '-':
		dotNext := c.runeAt(i.west()) == '.' || c.runeAt(i.east()) == '.'
		hyphenNext := c.runeAt(i.west()) == '\'' || c.runeAt(i.east()) == '\''
		return dotNext || hyphenNext

	case '|':
		dotAbove := c.runeAt(i.nWest()) == '.' || c.runeAt(i.nEast()) == '.'
		hyphenBelow := c.runeAt(i.sWest()) == '\'' || c.runeAt(i.sEast()) == '\''
		return dotAbove || hyphenBelow
	}

	return false
}

func (c *Canvas) withinBounds(i Index) bool {
	return i.x >= 0 && i.x < c.Width && i.y >= 0 && i.y < c.Height
}
