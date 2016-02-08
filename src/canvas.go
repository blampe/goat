package goaat

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

var JOINTS = []rune{'.', '\'', '+'}

func contains(in []rune, r rune) bool {
	for _, v := range in {
		if r == v {
			return true
		}
	}
	return false
}

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
			buffer.WriteRune(c.runeAt(idx))
		}
		buffer.WriteByte('\n')
	}

	return buffer.String()
}

func (c *Canvas) runeAt(i Index) rune {

	if val, ok := c.data[i]; ok {
		return val
	}

	return ' '
}

func NewCanvas(in io.Reader) Canvas {
	width := 0
	height := 0

	scanner := bufio.NewScanner(in)

	data := make(map[Index]rune)

	for scanner.Scan() {
		line := scanner.Text()

		for w := 0; w < len(line); w++ {
			idx := Index{x: w, y: height}
			data[idx] = rune(line[w])
		}

		if len(line) > width {
			width = len(line)
		}
		height++
	}

	return Canvas{Width: width, Height: height, data: data}
}

type Drawable interface {
	Draw(out io.Writer)
}

type Line struct {
	start  Index
	stop   Index
	dashed bool

	state lineState
}

type lineState int

const (
	EMPTY lineState = iota
	STARTED
	ENDED
)

func (l *Line) started() bool {
	return l.state != EMPTY
}

func (l *Line) ended() bool {
	return l.state == ENDED
}

func (l *Line) setStart(i Index) {
	if l.state == EMPTY {
		l.start = i
		l.state = STARTED
	}
}

func (l *Line) setStop(i Index) {
	if l.state == STARTED || l.state == ENDED {
		l.stop = i
		l.state = ENDED
	}
}

type Triangle struct {
	start       Index
	orientation Orientation
	halfStep    bool
}

type Circle struct {
	start Index
	bold  bool
}

type RoundedCorner struct {
	start       Index
	orientation Orientation
}

type Text struct {
	start    Index
	contents string
}

type Bridge struct {
	start       Index
	orientation Orientation
}

type Orientation int

const (
	NONE Orientation = iota
	N
	NE
	NW
	S
	SE
	SW
	E
	W
)

func (c *Canvas) Lines() []Line {
	var lines []Line

	lines = append(lines, c.linesFromIterator(upDown, []rune{'|', 'v', '^'})...)
	lines = append(lines, c.linesFromIterator(leftRight, []rune{'-', '<', '>', '(', ')'})...)
	lines = append(lines, c.linesFromIterator(diagUp, []rune{'/'})...)
	lines = append(lines, c.linesFromIterator(diagDown, []rune{'\\'})...)

	return lines
}

func (c *Canvas) linesFromIterator(ci canvasIterator, keepers []rune) []Line {
	var lines []Line

	var currentLine Line
	var lastSeenRune rune

	endCurrentLine := func(i Index) Line {
		if !currentLine.started() {
			currentLine.setStart(i)
		}
		if !currentLine.ended() {
			currentLine.setStop(currentLine.start)
			//currentLine.setStop(i)
		}
		lines = append(lines, currentLine)
		return Line{}
	}

	for idx := range ci(c.Width, c.Height) {
		r := c.runeAt(idx)

		isJoint := contains(JOINTS, r)
		shouldKeep := r != ' ' && (contains(keepers, r) || isJoint)

		if !shouldKeep && !currentLine.started() {
			continue
		}

		// Don't connect corner joints during diagonal sweeps, e.g:
		//    |
		//    +--
		// --+
		//   |

		if isJoint && contains(JOINTS, lastSeenRune) && (contains(keepers, '/') || contains(keepers, '\\')) {
			currentLine = endCurrentLine(idx)
			// Start a new line at this joint.
			currentLine.setStart(idx)
			lastSeenRune = r
			continue
		}

		notRoundedCorner := (!isJoint || c.isRoundedCorner(idx) == NONE)

		if !currentLine.started() && shouldKeep && notRoundedCorner {
			currentLine.setStart(idx)
		} else if currentLine.started() && shouldKeep && notRoundedCorner {
			currentLine.setStop(idx)
		}

		if !shouldKeep && currentLine.started() {
			currentLine = endCurrentLine(idx)
		}

		lastSeenRune = r
	}

	return lines
}

func (c *Canvas) Triangles() []Triangle {
	var triangles []Triangle

	o := NONE

	for idx := range upDown(c.Width, c.Height) {
		halfStep := false

		if c.isText(idx) {
			continue
		}

		r := c.runeAt(idx)

		switch r {
		case '^':
			o = N
			if c.runeAt(Index{idx.x, idx.y - 1}) == '-' {
				halfStep = true
			}
		case 'v':
			o = S
			if c.runeAt(Index{idx.x, idx.y + 1}) == '-' {
				halfStep = true
			}
		case '<':
			o = W
			if c.runeAt(Index{idx.x - 1, idx.y}) == '|' {
				halfStep = true
			}
		case '>':
			o = E
			if c.runeAt(Index{idx.x + 1, idx.y}) == '|' {
				halfStep = true
			}
		default:
			continue
		}

		triangles = append(triangles, Triangle{start: idx, orientation: o, halfStep: halfStep})
	}

	return triangles
}

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

func (c *Canvas) RoundedCorners() []RoundedCorner {
	var corners []RoundedCorner

	for idx := range leftRight(c.Width, c.Height) {
		if o := c.isRoundedCorner(idx); o != NONE {
			corners = append(corners, RoundedCorner{start: idx, orientation: o})
		}
	}

	return corners
}

// TODO foo
func (c *Canvas) isRoundedCorner(i Index) Orientation {

	r := c.runeAt(i)

	if r != '.' && r != '\'' {
		return NONE
	}

	left := Index{i.x - 1, i.y}
	right := Index{i.x + 1, i.y}
	lowerLeft := Index{i.x - 1, i.y + 1}
	lowerRight := Index{i.x + 1, i.y + 1}
	upperLeft := Index{i.x - 1, i.y - 1}
	upperRight := Index{i.x + 1, i.y - 1}

	if r == '.' {
		// North case

		//  .-
		// |
		if c.runeAt(right) == '-' && c.runeAt(lowerLeft) == '|' {
			return NW
		}

		// -.
		//   |
		if c.runeAt(left) == '-' && c.runeAt(lowerRight) == '|' {
			return NE
		}

	} else {
		// South case

		//   |
		// -'
		if c.runeAt(left) == '-' && c.runeAt(upperRight) == '|' {
			return SE
		}

		// |
		//  '-
		if c.runeAt(right) == '-' && c.runeAt(upperLeft) == '|' {
			return SW

		}

	}

	return NONE
}

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

	left := c.runeAt(Index{i.x - 1, i.y})
	right := c.runeAt(Index{i.x + 1, i.y})

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

	// If o or * and letters immediately adjacent, true.
	if (c.isTextLeft(i, 2) || c.isTextRight(i, 2)) && !c.hasIncomingLine(i) {
		return true
	}

	return false
}

func (c *Canvas) isTextLeft(i Index, limit uint8) bool {
	if limit == 0 {
		return false
	}
	left := Index{i.x - 1, i.y}

	return c.isDefinitelyText(left) || c.isTextLeft(left, limit-1)
}

func (c *Canvas) isTextRight(i Index, limit uint8) bool {
	if limit == 0 {
		return false
	}
	right := Index{i.x + 1, i.y}

	return c.isDefinitelyText(right) || c.isTextRight(right, limit-1)
}

func (c *Canvas) isDefinitelyText(i Index) bool {
	r := c.runeAt(i)

	if r == ' ' {
		return false
	}

	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}

	// Reserved characters
	if contains([]rune{'-', '|', '/', '\\', '^', 'v', '<', '>', '.', '\'', '(', ')', '+', 'o', '*'}, r) {
		// Check if we have any other reserved characters above us?
		// \|/
		// -x-
		// /|\
		return false
	}

	return true
}

func (c *Canvas) hasIncomingLine(i Index) bool {
	r := c.runeAt(i)

	north := Index{i.x, i.y - 1}
	south := Index{i.x, i.y + 1}
	east := Index{i.x + 1, i.y}
	west := Index{i.x - 1, i.y}
	nWest := Index{i.x - 1, i.y - 1}
	nEast := Index{i.x + 1, i.y - 1}
	sWest := Index{i.x - 1, i.y + 1}
	sEast := Index{i.x + 1, i.y + 1}

	switch r {
	case '*', 'o', '+':
		if c.runeAt(nWest) == '\\' || c.runeAt(north) == '|' || c.runeAt(nEast) == '/' ||
			c.runeAt(sWest) == '/' || c.runeAt(south) == '|' || c.runeAt(sEast) == '\\' {
			return true
		}

	case '|':
		if c.runeAt(north) == '|' ||
			c.runeAt(south) == '|' ||
			c.runeAt(north) == '\'' ||
			c.runeAt(north) == '.' ||
			c.runeAt(south) == '\'' ||
			c.runeAt(south) == '.' ||
			c.runeAt(nWest) == '.' ||
			c.runeAt(nEast) == '.' ||
			c.runeAt(sWest) == '\'' ||
			c.runeAt(sEast) == '\'' {
			return true
		}
	case '/':
		if c.runeAt(nEast) == '/' || c.runeAt(sWest) == '/' || contains(JOINTS, c.runeAt(nEast)) || contains(JOINTS, c.runeAt(sWest)) {
			return true
		}
	case '-':
		if c.runeAt(east) == '.' || c.runeAt(west) == '.' || c.runeAt(east) == '\'' || c.runeAt(west) == '\'' {
			return true
		}
	case '(', ')':
		if c.runeAt(north) == '|' || c.runeAt(south) == '|' {
			return true
		}
	}

	return false

}
