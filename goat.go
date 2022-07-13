/*
Package goat formats "ASCII-art" drawings into Github-flavored Markdown.

 <goat>
 porcelain API
                            BuildAndWriteSVG()
                               .----------.
     ASCII-art                |            |                      Markdown
      ----------------------->|            +------------------------->
                              |            |
                               '----------'
   · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · ·
 plumbing API

                                Canvas{}
               NewCanvas() .-------------------.  WriteSVGBody()
                           |                   |    .-------.
     ASCII-art    .--.     | data map[x,y]rune |   |  SVG{}  |    Markdown
      ---------->|    +--->| text map[x,y]rune +-->|         +------->
                  '--'     |                   |   |         |
                           '-------------------'    '-------'
 </goat>
*/
package goat

import (
	"bytes"
	"io"
)

// BuildAndWriteSVG reads in a newline-delimited ASCII diagram from src and writes a
// corresponding SVG diagram to dst.
func BuildAndWriteSVG(src io.Reader, dst io.Writer,
	svgColorLightScheme, svgColorDarkScheme string) {
	svg := buildSVG(src)
	writeBytes(dst, svg.String(svgColorLightScheme, svgColorDarkScheme))
}

func buildSVG(src io.Reader) SVG {
	var buff bytes.Buffer
	canvas := NewCanvas(src)
	canvas.WriteSVGBody(&buff)
	return SVG{
		Body:	buff.String(),
		Width:	canvas.widthScreen(),
		Height: canvas.heightScreen(),
	}
}
