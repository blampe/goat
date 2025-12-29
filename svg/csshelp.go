package svg

import (
	"fmt"
)

// generate sharable CSS, for use in simple diagrams
func ColorsOnlyCssFileContent(
	svgColorLightScheme, svgColorDarkScheme string) string {

	return fmt.Sprintf(
`    svg {
        fill: currentColor; stroke: currentColor;
        color: %s;  /* set value of 'currentColor' */
    }
    @media (prefers-color-scheme: dark) {
        svg {
            color: %s;  /* set value of 'currentColor' */
        }
    }
`,
		svgColorLightScheme,
		svgColorDarkScheme)
}


// Embed arg 'source' in the neo-attribute 'goat-source', which the browser merely
// ignores, except that it is displayed in the web debugger, for easy reference by the developer.
func newStyleElement(sourceOrigin, css string) string {
	return fmt.Sprintf(
`  <style type="text/css" source-text-origin="%s">
%s  </style>
`,
		sourceOrigin,
		css)
}

const defaultCSS = `    svg {
        color-scheme: light dark; /* this becomes necessary if not inherited from parent elements */
        font-family: monospace;
        font-size: 15px;
    }
    polyline, path {
        fill: none;
    }

    text {
        stroke: none;
        text-anchor: middle;
    }

    circle.filled {
        fill: inherit;
    }
    circle.hollow {
         fill: none;
    }
`
