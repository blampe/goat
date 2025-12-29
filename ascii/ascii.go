/*
  Format ASCII─art into SVG image files.
*/
package ascii

import (
	"io"

	"github.com/blampe/goat"
	"github.com/blampe/goat/svg"
)

// X  Differs from utf8.Canvas by the methods bound to it (see below).
type Canvas struct {
	svg.CanvasCommon
}

func (ac *Canvas) GetCommon() *svg.CanvasCommon {
	return &ac.CanvasCommon
}

func NewCanvas(config *svg.Config, in io.Reader) svg.AbstractCanvas {
	c := Canvas{
		CanvasCommon: svg.NewCanvasCommon(config, in),
	}
	// Fill the 'TextRunes' map, with runes removed from 'data', according to c.ShouldMoveToTextRunes()
	svg.MoveToText(&c)
	return &c
}

var verticalRunes = goat.MakeRuneSet(
	'|',   // VERTICAL LINE
)
var horizontalRunes = goat.MakeRuneSet(
	'-',   // HYPHEN
)

// XX  A/K/A "triangles"
var verticalArrowheadRunes = goat.MakeRuneSet(
	'v',
	'^',
)
var horizontalArrowheadRunes = goat.MakeRuneSet(
	'<',
	'>',
)
var arrowheadRunes = goat.UnionSets(
	verticalArrowheadRunes,
	horizontalArrowheadRunes,
)

// Characters where more than one line segment can come together.
var jointRunes = goat.MakeRuneSet(
		'.',     // possible ...    top corner of a 90 degree angle, or curve
		'\'',    // possible ... bottom corner of a 90 degree angle, or curve
		'+',
		'*',
		'o',
	)

// XX  'Reserved' is not a faithful abstraction; the functional meaning is more
//     like "possibly graphical, depending on neighbors".
//      X  All but ' ' below might be well called "pathRunes".
var ReservedSet = goat.UnionSets(
	jointRunes, goat.MakeRuneSet(
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
		' ',   // X SPACE is "reserved"
	))

// "wide" implies that characters to left and right must not be text, if the "wide" character is
// to be rendered as graphics
var wideSVGSet = goat.MakeRuneSet(
	// double-wide
	'o',
	'*',

	// "single-wide"
	'v',   // X  Input containing " over " needs to be considered text.
//	'>',   // Uncommenting would get 'o<' and '>o' wrong.  But o> and >o -- never desired to be text?
//	'<',   // Uncom...
	'^',
	')',
	'(',
	'.',   // Dropping this would cause " v. " to be considered graphics.
)


// XX  rename 'isSpot()'?
func isDot(r rune) bool {
	return r == 'o' || r == '*'
}

func isJoint(r rune) bool {
	return jointRunes.Contains(r)
}

func isTriangle(r rune) bool {
	return arrowheadRunes.Contains(r)
}

func (c *Canvas) ShouldMoveToTextRunes(i svg.XyIndex) bool {
	i_r := c.RuneAt(i)
	// character := string(i_r); _ = character   // for debug

	if _, found := ReservedSet[i_r]; !found {
		return true
	}

	// X  After this point, deal with problematic cases of 'reserved' characters that
	//    nevertheless must be treated as ordinary text.

	// This is a reserved character with an incoming line (e.g., "|") above or below it,
	// so call it non-text.
	if c.haslineAboveOrBelow(i) {
		return false
	}

	// Returns true if index 'i' of c.Data[] is to be treated as reserved.
	// X  Characters like 'o' and 'v' need more context (e.g., are other text characters
	//    nearby) to determine whether they're part of a diagram.
	isReserved := func(i svg.XyIndex) (found bool) {
		i_r, inData := c.Data[i]
		if !inData {
			// lies off left or right end of line, treat as reserved
			return true
		}
		_, found = ReservedSet[i_r]
		return
	}

	left := i.West()
	right := i.East()

	// XX ? Generalize these two tests to regard arbitrarily long rows of reserved chars as
	//      text, if even a single non-reserved character is embedded in the row.
	//  X  For test cases, see regression.txt.
	if !isReserved(left) {
		return true
	}
	if !isReserved(right) {
		return true
	}

	crowded := func (l, r svg.XyIndex) bool {
		return  svg.InSet(wideSVGSet, c.Data, l) &&
			svg.InSet(wideSVGSet, c.Data, r)
	}
	if crowded(left, i) || crowded(i, right) {
		return true
	}

	// If 'i' has anything other than a space to either left or right, treat as non-text.
	if !(c.RuneAt(left) == ' ' && c.RuneAt(right) == ' ') {
		return false
	}

	// circles surrounded by whitespace shouldn't be shown as text.
	if i_r == 'o' || i_r == '*' {
		return false
	}

	// 'i' is surrounded by whitespace or text on one side or the other, at two cell's distance.
	// XX  Cause of superfluous space-containing <text> elements (two!) at each end of strings?    
	if !isReserved(left.West()) || !isReserved(right.East()) {
		return true
	}

	return false
}


// Returns true if it looks like this character belongs to anything besides a
// horizontal line. This is the context we use to determine if a reserved
// character is text or not.
func (c *Canvas) haslineAboveOrBelow(i svg.XyIndex) bool {
	i_r := c.RuneAt(i)

	switch i_r {
	case '*', 'o', '+', 'v', '^':
		return c.partOfDiagonalline(i) || c.partOfVerticalline(i)
	case '|':
		return c.partOfVerticalline(i) || c.partOfroundedCorner(i)
	case '/', '\\':
		return c.partOfDiagonalline(i)
	case '-':
		return c.partOfroundedCorner(i)
	case '(', ')':
		return c.partOfVerticalline(i)
	}

	return false
}

// Returns true if a "|" segment passes through this index.
func (c *Canvas) partOfVerticalline(i svg.XyIndex) bool {
	this := c.RuneAt(i)
	north := c.RuneAt(i.North())
	south := c.RuneAt(i.South())

	jointAboveMe := verticalRunes.Contains(this) && isJoint(north)

	if verticalRunes.Contains(north) || jointAboveMe {
		return true
	}

	jointBelowMe := verticalRunes.Contains(this) && isJoint(south)

	if verticalRunes.Contains(south) || jointBelowMe {
		return true
	}

	return false
}

// Return true if a "--" segment passes through this index.
func (c *Canvas) partOfHorizontalline(i svg.XyIndex) bool {
	return  horizontalRunes.Contains(c.RuneAt(i.East())) ||
		horizontalRunes.Contains(c.RuneAt(i.West()))
}

func (c *Canvas) partOfDiagonalline(i svg.XyIndex) bool {
	r := c.RuneAt(i)

	n := c.RuneAt(i.North())
	s := c.RuneAt(i.South())
	nw := c.RuneAt(i.NWest())
	se := c.RuneAt(i.SEast())
	ne := c.RuneAt(i.NEast())
	sw := c.RuneAt(i.SWest())

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
func (c *Canvas) partOfroundedCorner(i svg.XyIndex) bool {
	r := c.RuneAt(i)

	switch r {
	case '-':
		dotNext := c.RuneAt(i.West()) == '.' || c.RuneAt(i.East()) == '.'
		hyphenNext := c.RuneAt(i.West()) == '\'' || c.RuneAt(i.East()) == '\''
		return dotNext || hyphenNext

	case '|':
		dotAbove := c.RuneAt(i.NWest()) == '.' || c.RuneAt(i.NEast()) == '.'
		hyphenBelow := c.RuneAt(i.SWest()) == '\'' || c.RuneAt(i.SEast()) == '\''
		return dotAbove || hyphenBelow
	}

	return false
}

// Should the output for this XyIndex be a half-height line segment, and if so,
// a top-half 'O_N' or a bottom-half 'O_S'?
// TODO: Have this take care of all the vertical line nudging.
func (c *Canvas) partOfHalfStep(i svg.XyIndex) svg.Orientation {
	r := c.RuneAt(i)
	if r != '\'' && r != '.' && r != '|' {
		return svg.O_NONE
	}

	if c.isroundedCorner(i) != svg.O_NONE {
		return svg.O_NONE
	}

	w := c.RuneAt(i.West())
	e := c.RuneAt(i.East())
	n := c.RuneAt(i.North())
	s := c.RuneAt(i.South())
	nw := c.RuneAt(i.NWest())
	ne := c.RuneAt(i.NEast())

	switch r {
	case '\'':
		//  _	   _
		//   '-	 -'
		if (nw == '_' && e == '-') || (w == '-' && ne == '_') {
			return svg.O_N
		}
	case '.':
		// _.-	-._
		if (w == '-' && e == '_') || (w == '_' && e == '-') {
			return svg.O_S
		}
	case '|':
		//// _	 _
		////  | |
		if n != '|' && (ne == '_' || nw == '_') {
			return svg.O_N
		}

		if n == '-' {
			return svg.O_N
		}

		//// _| |_
		if s != '|' && (w == '_' || e == '_') {
			return svg.O_S
		}

		if s == '-' {
			return svg.O_S
		}
	}
	return svg.O_NONE
}
