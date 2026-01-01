package utf8

import (
	"github.com/blampe/goat"
	"github.com/blampe/goat/svg"
)

// X  Fatal error on unsupported BOX characters, rather than treating
//    simply as text?

// X  Should "dangling" BOX DRAWINGS ends be supported in SVG output?
//      => Principle of 'least surprise': allow, as normal case.

// What about 'BOX DRAWINGS HEAVY *'?  => no rounded corners available

var squareCornerRunes = goat.MakeRuneSet(
	//        Orientation
	//        |          abutment with Line's
	//        |          |          |
	'┌',  //  svg.O_NW:  O_E start  O_S start
	'└',  //  svg.O_SW:  O_E start  O_S stop
	'┐',  //  svg.O_NE:  O_E stop   O_S start
	'┘',  //  svg.O_SE:  O_E stop   O_S stop
)

// BOX DRAWINGS LIGHT ARC characters
var roundedCornerRunes = goat.MakeRuneSet(
	//        Orientation
	//        |          abutment with Line's
	//        |          |          |
	'╭',  //  svg.O_NW:  O_E start  O_S start
	'╰',  //  svg.O_SW:  O_E start  O_S stop
	'╮',  //  svg.O_NE:  O_E stop   O_S start
	'╯',  //  svg.O_SE:  O_E stop   O_S stop
)

var roundedCornerCenters = map[svg.Orientation]svg.Pixel{
	svg.O_NW: { X: cornerRadius, Y: cornerRadius},   // '╭'
	svg.O_SW: { X: cornerRadius, Y:-cornerRadius},   // '╰'
	svg.O_NE: { X:-cornerRadius, Y: cornerRadius},   // '╮'
	svg.O_SE: { X:-cornerRadius, Y:-cornerRadius},   // '╯'
}
func (c *Canvas) CenterPixel(rc svg.RoundedCorner) svg.Pixel {
	cellCenter := rc.Start.AsPixel()
	return cellCenter.Sum(roundedCornerCenters[rc.Orientation])
}

var boxJointRunes = goat.UnionSets(
	squareCornerRunes,
	roundedCornerRunes,
	goat.MakeRuneSet(
		'┬',
		'┴',
		'┤',
		'├',
		'┼',
	),
)

// Meaning is "draw a Line, possibly extending into the adjacent cell
// lying on side 'Orientation'."
var connects = make(map[svg.Orientation]goat.RuneSet)

func init() {
	connects[svg.O_E] = goat.MakeRuneSet(
		'╶',
		'─',   // BOX DRAWINGS LIGHT HORIZONTAL
		'╭',  // BOX DRAWINGS LIGHT ARC
		'╰',
		'┌',
		'└',
		'┬',
		'┴',
		'├',
		'┼')
	connects[svg.O_W] = goat.MakeRuneSet(
		'╴',
		'─',
		'╮',
		'╯',
		'┐',
		'┘',
		'┬',
		'┴',
		'┤',
		'┼')

	connects[svg.O_S] = goat.MakeRuneSet(
		'╷',
		'│',   // BOX DRAWINGS LIGHT VERTICAL
		'╮',
		'╭',
		'┌',
		'┐',
		'┬',
		'┤',
		'├',
		'┼')
	connects[svg.O_N] = goat.MakeRuneSet(
		'╵',
		'│',
		'╯',
		'╰',
		'└',
		'┘',
		'┴',
		'┤',
		'├',
		'┼')
}
