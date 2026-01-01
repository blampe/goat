package svg

import (
	"github.com/blampe/goat"
)

// XyIndex represents a position within an ASCII diagram, and
// of the visual center of a corresponding rectangle within the output SVG,
// where both have a height:width ratio of 2:1.
type XyIndex struct {
	// units of cells
	X, Y int
}


// Arg 'canvasMap' is typically either Canvas.data or Canvas.text
func InSet(set goat.RuneSet, canvasMap map[XyIndex]rune, i XyIndex) bool {
	r, inMap := canvasMap[i]
	if !inMap {
		return false 	// r == rune(0)
	}
	return set.Contains(r)
}


// Type "pixel' represents the CSS-pixel coordinates of the apparent visual center of
// an 8x16 cell pointed to by an XyIndex.
type Pixel struct {
	// units of CSS "pixels"
	X, Y int
}

func (a *Pixel) Delta(b Pixel) {
	a.X += b.X
	a.Y += b.Y
}

func (a Pixel) Sum(b Pixel) Pixel {
	return Pixel{
		a.X + b.X,
		a.Y + b.Y,
	}
}

const (
	CellWidth = 8
	CellHeight = 16
	W = CellWidth
	H = CellHeight
)


func (i *XyIndex) AsPixel() Pixel {
	// TODO  define constants rather than hard-wire width and height of cell
	return Pixel{
		X: i.X * CellWidth,
		Y: i.Y * CellHeight}
}

func (i *XyIndex) AsPixelXY() (int, int) {
	p := i.AsPixel()
	return p.X, p.Y
}

func (i *XyIndex) East() XyIndex {
	return XyIndex{i.X + 1, i.Y}
}

func (i *XyIndex) West() XyIndex {
	return XyIndex{i.X - 1, i.Y}
}

func (i *XyIndex) North() XyIndex {
	return XyIndex{i.X, i.Y - 1}
}

func (i *XyIndex) South() XyIndex {
	return XyIndex{i.X, i.Y + 1}
}

func (i *XyIndex) NWest() XyIndex {
	return XyIndex{i.X - 1, i.Y - 1}
}

func (i *XyIndex) NEast() XyIndex {
	return XyIndex{i.X + 1, i.Y - 1}
}

func (i *XyIndex) SWest() XyIndex {
	return XyIndex{i.X - 1, i.Y + 1}
}

func (i *XyIndex) SEast() XyIndex {
	return XyIndex{i.X + 1, i.Y + 1}
}
