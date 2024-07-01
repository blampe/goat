package goat

import (
	"bufio"
	"io"
)

type (
	exists struct{}
	runeSet map[rune]exists
)

// Characters where more than one line segment can come together.
var jointRunes = []rune{
	'.',     // possible ...    top corner of a 90 degree angle, or curve
	'\'',    // possible ... bottom corner of a 90 degree angle, or curve
	'+',
	'*',
	'o',
}

var reserved = append(
	jointRunes,
	[]rune{
		'-',
		'_',
		'|',
		'v',
		'^',
		'>',
		'<',
		'/',
		'\\',
		')',
		'(',
		' ',   // X SPACE is reserved
	}...,
)
var reservedSet runeSet

var doubleWideSVG = []rune{
	'o',
	'*',
}
var wideSVG = []rune{
	'v',   // X  Input containing " over " needs to be considered text.
//	'>',   // Uncommenting would get 'o<' and '>o' wrong.  But o> and >o -- never desired to be text?
//	'<',   // ibid.
	'^',
	')',
	'(',
	'.',   // Dropping this would cause " v. " to be considered graphics.
}
var wideSVGSet = makeSet(append(doubleWideSVG, wideSVG...))

func makeSet(runeSlice []rune) (rs runeSet) {
	rs = make(runeSet)
	for _, r := range runeSlice {
		rs[r] = exists{}
	}
	return
}

func init() {
	// Recall that ranging over a 'string' type extracts values of type 'rune'.

	reservedSet = make(runeSet)
	for _, r := range reserved {
		reservedSet[r] = exists{}
	}
}

// XX  linear search of slice -- alternative to a map test
func contains(in []rune, r rune) bool {
	for _, v := range in {
		if r == v {
			return true
		}
	}
	return false
}

func isJoint(r rune) bool {
	return contains(jointRunes, r)
}

// XX  rename 'isSpot()'?
func isDot(r rune) bool {
	return r == 'o' || r == '*'
}

func isTriangle(r rune) bool {
	return r == '^' || r == 'v' || r == '<' || r == '>'
}

// Arg 'canvasMap' is typically either Canvas.data or Canvas.text
func inSet(set runeSet, canvasMap map[Index]rune, i Index) (inset bool) {
	r, inMap := canvasMap[i]
	if !inMap {
		return false 	// r == rune(0)
	}
	_, inset = set[r]
	return
}

// Looks only at c.data[], ignores c.text[].
// Returns the rune for ASCII Space i.e. ' ', in the event that map lookup fails.
//  XX  Name 'dataRuneAt()' would be more descriptive, but maybe too bulky.
func (c *Canvas) runeAt(i Index) rune {
	if val, ok := c.data[i]; ok {
		return val
	}
	return ' '
}

// Canvas represents a 2D ASCII rectangle.
type Canvas struct {
	// units of cells
	Width, Height int

	data   map[Index]rune
	text   map[Index]rune
}

func (c *Canvas) heightScreen() int {
	// XX  Why " + 8 + 1"?
	return c.Height*16 + 8 + 1
}

func (c *Canvas) widthScreen() int {
	// XX  Why "c.Width + 1"?
	return (c.Width + 1) * 8
}


// NewCanvas creates a fully-populated Canvas according to GoAT-formatted text read from
// an io.Reader, consuming all bytes available.
func NewCanvas(in io.Reader) (c Canvas) {
	//  XX  Move this function to top of file.
	width := 0
	height := 0

	scanner := bufio.NewScanner(in)

	c = Canvas{
		data:	make(map[Index]rune),
		text:	nil,
	}

	// Fill the 'data' map.
	for scanner.Scan() {
		lineStr := scanner.Text()

		w := 0
		// X  Type of second value assigned from "for ... range" operator over a string is "rune".
		//               https://go.dev/ref/spec#For_statements
		//    But yet, counterintuitively, type of lineStr[_index_] is 'byte'.
		//               https://go.dev/ref/spec#String_types
		// XXXX  Refactor to use []rune from above.
		for _, r := range lineStr {
			//if r > 255 {
			//	fmt.Printf("linestr=\"%s\"\n", lineStr)
			//	fmt.Printf("r == 0x%x\n", r)
			//}
			if r == '	' {
				panic("TAB character found on input")
			}
			i := Index{w, height}
			c.data[i] = r
			w++
		}

		if w > width {
			width = w
		}
		height++
	}

	c.Width = width
	c.Height = height
	c.text = make(map[Index]rune)
	// Fill the 'text' map, with runes removed from 'data'.
	c.MoveToText()
	return
}

// Move contents of every cell that appears, according to a tricky set of rules,
// to be "text", into a separate map: from data[] to text[].
// So data[] and text[] are an exact partitioning of the
// incoming grid-aligned runes.
func (c *Canvas) MoveToText() {
	for i := range leftRight(c.Width, c.Height) {
		if c.shouldMoveToText(i) {
			c.text[i] = c.runeAt(i)	// c.runeAt() Reads from c.data[]
		}
	}
	for i := range c.text {
		delete(c.data, i)
	}
}


func (c *Canvas) shouldMoveToText(i Index) bool {
	i_r := c.runeAt(i)
	if i_r == ' ' {
		// X  Note that c.runeAt(i) returns ' ' if i lies right of all chars on line i.Y
		return false
	}

	// Returns true if the character at index 'i' of c.data[] is reserved for diagrams.
	// Characters like 'o' and 'v' need more context (e.g., are other text characters
	// nearby) to determine whether they're part of a diagram.
	isReserved := func(i Index) (found bool) {
		i_r, inData := c.data[i]
		if !inData {
			// lies off left or right end of line, treat as reserved
			return true
		}
		_, found = reservedSet[i_r]
		return
	}

	if !isReserved(i) {
		return true
	}

	// This is a reserved character with an incoming line (e.g., "|") above or below it,
	// so call it non-text.
	if c.hasLineAboveOrBelow(i) {
		return false
	}

	w := i.west()
	e := i.east()

	// Reserved characters like "o" or "*" with letters sitting next to them
	// are probably text.
	// TODO: Fix this to count contiguous blocks of text. If we had a bunch of
	// reserved characters previously that were counted as text then this
	// should be as well, e.g., "A----B".

	// 'i' is reserved but surrounded by text and probably part of an existing word.
	// Preserve chains of reserved-but-text characters like "foo----bar".
	if textLeft := !isReserved(w); textLeft {
		return true
	}
	if textRight := !isReserved(e); textRight {
		return true
	}

	crowded := func (l, r Index) bool {
		return  inSet(wideSVGSet, c.data, l) &&
			inSet(wideSVGSet, c.data, r)
	}
	if crowded(w, i) || crowded(i, e) {
		return true
	}

	// If 'i' has anything other than a space to either left or right, treat as non-text.
	if !(c.runeAt(w) == ' ' && c.runeAt(e) == ' ') {
		return false
	}

	// Circles surrounded by whitespace shouldn't be shown as text.
	if i_r == 'o' || i_r == '*' {
		return false
	}

	// 'i' is surrounded by whitespace or text on one side or the other, at two cell's distance.
	if !isReserved(w.west()) || !isReserved(e.east()) {
		return true
	}

	return false
}


// Returns true if it looks like this character belongs to anything besides a
// horizontal line. This is the context we use to determine if a reserved
// character is text or not.
func (c *Canvas) hasLineAboveOrBelow(i Index) bool {
	i_r := c.runeAt(i)

	switch i_r {
	case '*', 'o', '+', 'v', '^':
		return c.partOfDiagonalLine(i) || c.partOfVerticalLine(i)
	case '|':
		return c.partOfVerticalLine(i) || c.partOfRoundedCorner(i)
	case '/', '\\':
		return c.partOfDiagonalLine(i)
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

	jointAboveMe := this == '|' && isJoint(north)

	if north == '|' || jointAboveMe {
		return true
	}

	jointBelowMe := this == '|' && isJoint(south)

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
	r := c.runeAt(i)

	n := c.runeAt(i.north())
	s := c.runeAt(i.south())
	nw := c.runeAt(i.nWest())
	se := c.runeAt(i.sEast())
	ne := c.runeAt(i.nEast())
	sw := c.runeAt(i.sWest())

	switch r {
	// Diagonal segments can be connected to joint or other segments.
	case '/':
		return ne == r || sw == r || isJoint(ne) || isJoint(sw) || n == '\\' || s == '\\'
	case '\\':
		return nw == r || se == r || isJoint(nw) || isJoint(se) || n == '/' || s == '/'

	// For everything else just check if we have segments next to us.
	default:
		return nw == '\\' || ne == '/' || sw == '/' || se == '\\'
	}
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

// TODO: Have this take care of all the vertical line nudging.
func (c *Canvas) partOfHalfStep(i Index) Orientation {
	r := c.runeAt(i)
	if r != '\'' && r != '.' && r != '|' {
		return NONE
	}

	if c.isRoundedCorner(i) != NONE {
		return NONE
	}

	w := c.runeAt(i.west())
	e := c.runeAt(i.east())
	n := c.runeAt(i.north())
	s := c.runeAt(i.south())
	nw := c.runeAt(i.nWest())
	ne := c.runeAt(i.nEast())

	switch r {
	case '\'':
		//  _	   _
		//   '-	 -'
		if (nw == '_' && e == '-') || (w == '-' && ne == '_') {
			return N
		}
	case '.':
		// _.-	-._
		if (w == '-' && e == '_') || (w == '_' && e == '-') {
			return S
		}
	case '|':
		//// _	 _
		////  | |
		if n != '|' && (ne == '_' || nw == '_') {
			return N
		}

		if n == '-' {
			return N
		}

		//// _| |_
		if s != '|' && (w == '_' || e == '_') {
			return S
		}

		if s == '-' {
			return S
		}
	}
	return NONE
}
