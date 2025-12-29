package svg

import (
	"io"
)

// Drawable represents anything that can Draw itself.
type Drawable interface {
	Draw(out io.Writer)
}

// XX  drop names 'start' below

// Triangle corresponds to '^', 'v', '<' and '>' runes in the absence of
// surrounding alphanumerics.
type Triangle struct {
	Start	     XyIndex
	Orientation  Orientation
	NeedsNudging bool
}

// Circle corresponds to 'o' or '*' runes in the absence of surrounding
// alphanumerics.
type Circle struct {
	Start XyIndex
	Bold  bool
}

type RoundedCorner struct {
	Start	    XyIndex
	Orientation Orientation
}

// Bridge corresponds to combinations of "-)-" or "-(-" and is displayed as
// the vertical line "hopping over" the horizontal.
type Bridge struct {
	Start	    XyIndex
	Orientation Orientation
}

// Orientation represents the primary direction that a Drawable is facing.
type Orientation int

const (
	O_NONE Orientation = iota // No orientation; no structure present.
	O_N			// North
	O_NE			// Northeast
	O_NW			// Northwest
	O_S			// South
	O_SE			// Southeast
	O_SW			// Southwest
	O_E			// East
	O_W			// West
)
