package svg

import (
	"fmt"
)

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

// See:
//   https://drafts.csswg.org/mediaqueries-5/#prefers-color-scheme
//   https://developer.mozilla.org/en-US/docs/Web/SVG/Element/style
//   https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute
//
//   X   Note that all elements are drawn by SVG not HTML, and this guidance about
//       the CSS "color:" property is not valid:
//           https://developer.mozilla.org/en-US/docs/Web/CSS/color
//       Rather, the authority is:
//           https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute
//       Or
//           https://developer.mozilla.org/en-US/docs/Web/SVG/Reference/Element/text
//         containing this:
//           /* Note that the color of the text is set with the    *
//            * fill property, the color property is for HTML only */
//
//  XX  Could this be moved into "embed:defaultCSS", to be loaded unless CLI requests otherwise?  
const defaultCSS = `    svg {
        color-scheme: light dark; /* this becomes necessary if not inherited from parent elements */
        font-family: monospace;
        font-size: 15px;
    }
    text {
        stroke: none;
        text-anchor: middle;
    }
    .path {
        fill: none;
    }
    circle.filled {
        fill: inherit;
    }
    circle.hollow {
         fill: none;
    }
`

// generate sharable CSS, for use in simple diagrams
func ColorsOnlyCssFileContent(
	svgColorLightScheme, svgColorDarkScheme string) string {

	return fmt.Sprintf(
`    svg {
        color: %s;  /* set value of 'currentColor' */
    }
    @media (prefers-color-scheme: dark) {
        svg {
            color: %s;  /* set value of 'currentColor' */
        }
    }
    svg {
        fill: currentColor;
        stroke: currentColor;
    }
`,
		svgColorLightScheme,
		svgColorDarkScheme)
}
