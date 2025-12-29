/*
  Format Unicode BOX─art into SVG image files.
*/
package utf8

import (
	"io"

	"github.com/blampe/goat"
	"github.com/blampe/goat/svg"
)

// X  Differs from ascii.Canvas by the methods bound to it (see below).
type Canvas struct {
	// ? make explicitly a local member, requiring dot-qualification?
	//     XX  would require wrapper accessor methods for .Width, .Height, RuneAt().
	svg.CanvasCommon
}

func (ac *Canvas) GetCommon() *svg.CanvasCommon {
	return &ac.CanvasCommon
}

// Copy-paste of ascii.NewCanvas() -- difference is in data types utf8.Canvas vs ascii.Canvas
//    XX  simplify, by pushing call to svg.NewCanvasCommon() up to client?   
func NewCanvas(config *svg.Config, in io.Reader) svg.AbstractCanvas {
	c := Canvas{
		CanvasCommon: svg.NewCanvasCommon(config, in),
	}
	// Fill the 'TextRunes' map, with runes removed from 'data', according to c.ShouldMoveToTextRunes()
	svg.MoveToText(&c)
	return &c
}

var verticalRunes = goat.MakeRuneSet(
	'│',   // BOX DRAWINGS LIGHT VERTICAL
	'╷',
	'╵',
)
var horizontalRunes = goat.MakeRuneSet(
	'─',   // BOX DRAWINGS LIGHT HORIZONTAL
	'╶',
	'╴',
)

// A/K/A "triangles"
var verticalArrowheadRunes = goat.MakeRuneSet(
	'▼',  // ▼
	'▲',  // ▲
)
var leftArrowheadRunes = goat.MakeRuneSet(
	'◀',  //  ◀
	'◄',  //  ◄
)
var rightArrowheadRunes = goat.MakeRuneSet(
	'▶',  //  ▶
	'►',  //  ►
)
var horizontalArrowheadRunes = goat.UnionSets(
	leftArrowheadRunes,
	rightArrowheadRunes,
)
var arrowheadRunes = goat.UnionSets(
	verticalArrowheadRunes,
	horizontalArrowheadRunes,
)

// X  Parameterize the output SVG <circle> with CSS to create the variants.
var dotRunes = goat.MakeRuneSet(
	'●',  // ●  BLACK CIRCLE  0x25CF
	'○',  // ○  WHITE CIRCLE  0x25CB
	//'◌',  // ◌  X  requires <circle> with reference to a CSS class for the dot patterning
	//'◠',  // ◠  XX draw with <path>, containing a 180-degree arc.
	//'◡',  // ◡
)

// X  Parameterize the output SVG <rect> with CSS to create the variants.
var squareRunes = goat.MakeRuneSet(
	//'◼',  // ◼
	//'◻',  // ◻
	//'⬚',  // ⬚
	//'▢',  // ▢
	//'◾', // ◾
	//'▫',  // ▫
)
var boxEdgeRunes = goat.UnionSets(
	verticalRunes,
	horizontalRunes,
)
var ReservedSet = goat.UnionSets(
	boxJointRunes, boxEdgeRunes, arrowheadRunes, dotRunes,
	goat.MakeRuneSet(
		' ',   // X SPACE is "reserved"
	))

// XX  rename 'isCircle()'?
func isDot(r rune) bool {
	return dotRunes.Contains(r)
}

func (c *Canvas) ShouldMoveToTextRunes(i svg.XyIndex) bool {
	i_r := c.RuneAt(i)
	// character := string(i_r); _ = character   // for debug

	if _, found := ReservedSet[i_r]; !found {
		return true
	}
	return false
}
