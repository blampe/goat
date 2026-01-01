package ascii

import (
	"io"

	"github.com/blampe/goat/svg"
)

// XX  ? Collect all ASCII-specific drawing functions into one file?
func drawArc(out io.Writer, rc svg.RoundedCorner, radius int) {
	startPixel := rc.Start.AsPixel()
	var delta svg.Pixel

	switch rc.Orientation {
	case svg.O_NW:
		delta = svg.Pixel{ radius/2, radius}
	case svg.O_SW:
		delta = svg.Pixel{ radius/2,-radius}
	case svg.O_NE:
		delta = svg.Pixel{-radius/2, radius}
	case svg.O_SE:
		delta = svg.Pixel{-radius/2,-radius}
	}
	centerPixel := startPixel.Sum(delta)
	rc.DrawCentered(out, centerPixel, radius)
}
