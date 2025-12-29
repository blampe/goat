package svg

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/blampe/goat/internal"
)

type textDrawer struct {
	config *Config

	// Support nesting of SVG anchor elements <a> -- report input errors promptly. 
	wrapperStack []wrapper
}

type wrapper struct {
	*markBinding
	text
}

func Writetext(out io.Writer, config *Config, ac AbstractCanvas) {
	tD := textDrawer{
		config: config,
	}
	cc := ac.GetCommon()
	for _, textObj := range cc.text() {
		err := tD.Draw(out, textObj)
		if err != nil {
			// XX  Dump the filename as well.
			log.Fatalf(`
%s
In Config %+v`,
				err.Error(),
				config,
			)
		}
	}
	// diagnose bad remaining wrapperStack contents here
	// XX share error state dumping code with Draw() below.
	// XX XX  End users want helpful diagnostics.
	//         XX  Passing an 'error' return up would require changing 3-4 levels of callers.
	//         X   Alternative: set a field in Config 'exitStatus', to be checked by
	//             interested callers.
	if len(tD.wrapperStack) > 0 {
		wrapper := tD.wrapperStack[0]
		log.Fatalf(`
len(tD.wrapperStack)==%d, should be empty.

tD.wrapperStack[0] = {
  markBinding: %s
  text: %s
}
`,
			len(tD.wrapperStack),
			wrapper.markBinding.String(),
			wrapper.text.String(),
		)
	}
}

// Output is nested inside <svg>...</svg>.
// Draw a single text character as an SVG <text> element. (In HTML, there is no <text> element.)
// Also emit grouping element <a>, modifying effect of contained <text> elements.
//  XX For ease of debugging output in browser's Inspector, group by words, e.g. 
//       	var lastX int
//		if t.Start.X != lastX + 1 {
//			internal.MustFPrintf(out, "  </g>\n  <g>\n")
//		}
//		lastX = t.Start.X

func (tD *textDrawer) Draw(out io.Writer, t text) error {
	//character := string(t.r); _ = character   // for debug
	textIndex := t.Start
	p := textIndex.AsPixel()

	endMarkBinding, foundEndMark := tD.config.endMap[t.r]
	beginMarkBinding, foundBeginMark := tD.config.beginMap[t.r]

	handleBeginMark := func() {
		tD.wrapperStack = append(tD.wrapperStack, 
			wrapper{beginMarkBinding, t})

		var attrs string
		if len(beginMarkBinding.HRef) > 0 {
			// X  Observed in browser: empty string "href=''" produces linking to the page itself.
			//    Therefore, drop the href attribute entirely -- apparently functional equivalent
			//    of a <g> element.
			attrs += fmt.Sprintf(" href='%s'", beginMarkBinding.HRef)
		}
		if len(beginMarkBinding.ClassNames) > 0 {
			attrs += fmt.Sprintf(" class='%s", beginMarkBinding.ClassNames[0])
			for _, classname := range beginMarkBinding.ClassNames[1:] {
				attrs += fmt.Sprintf(" %s", classname)
			}
			attrs += fmt.Sprintf("'")
		}
		internal.MustFPrintf(out, "  <a%s>\n", attrs)
		//subst := beginMarkBinding.subst[0]
		//
		//if subst != 0 {
		//	finalDraw(out, p, subst)
		//}
	}

	if foundEndMark {
		elemString := "a"
		if len(tD.wrapperStack) == 0 {
			if foundBeginMark {
				handleBeginMark()
				return nil
			}
			return errors.New(fmt.Sprintf(
				// toggling case?
				"\n\tEnd mark %s for\n" +
					"\t\t markpairclass %+v found at\n" +
					"\t\tIndex %+v, but no matching start mark on stack",
				string(t.r), endMarkBinding, t.Start))
		}
		// Verify that t.r is the expected end mark.
		tosWrapper := tD.wrapperStack[len(tD.wrapperStack)-1]
		tosMarkBinding := tosWrapper.markBinding
		if t.r != tosMarkBinding.markpair[1] {
			// toggling case?
			if foundBeginMark {
				handleBeginMark()
				return nil
			}
			tosMarkBindingEndMark := tosMarkBinding.markpair[1]
			return errors.New(fmt.Sprintf(
				"\tAt line %d, column %d scanned unexpected end mark" +
					"\n\t%s (0x%x) for markpairclass" +
					"\n\t\t%s" +
					"\n\texpected end mark" +
					"\n\t%s (0x%x) for markpairclass" +
					"\n\t\t%s",
				textIndex.Y, textIndex.X,
				string(t.r), t.r,
				formatMarkBinding(endMarkBinding),
				string(tosMarkBindingEndMark), tosMarkBindingEndMark,
				formatMarkBinding(tosMarkBinding)))
		}
		//// Special case: do not show a styled SPACE (possibly underlined)
		//// as the replacement for an end Mark character.
		//subst := endMarkBinding.subst[1]
		//if subst != 0 {
		//	finalDraw(out, p, subst)
		//}

		internal.MustFPrintf(out, "</%s>\n", elemString)
		// pop the stack
		tD.wrapperStack = tD.wrapperStack[:len(tD.wrapperStack)-1]
		return nil
	}
	if foundBeginMark {
		handleBeginMark()
	} else {
		finalDraw(out, p, t.r)
	}
	return nil
}


func finalDraw(out io.Writer, p Pixel, r rune) {
	if r == 0 {
		log.Panicf("NULL rune received")
	}
	c := string(r)
	if len(c) <= 0 {
		log.Panicf("rune %#v yielded empty string!", r)
	}
	if c == " " {
		// <text> elements containing only a SPC character can always be safely dropped, yes?
		// XX  Could this case be eliminated earlier in processing?
		return
	}
	opacity := 0

	// Markdeep special-cases these characters and treats them like a
	// checkerboard.
	switch c {
	case "▉":
		opacity = -9999
	case "▓":
		opacity = 64
	case "▒":
		opacity = 128
	case "░":
		opacity = 191
	}

	if opacity != 0 {
		fill := "currentColor"
		if opacity > 0 {
			fill = fmt.Sprintf("rgb(%d,%d,%d)", opacity, opacity, opacity)
		}
		internal.MustFPrintf(out,
			`<rect x="%d" y="%d" width="%d" height="%d" fill="%s"></rect>
`,
			p.X-W/2, p.Y-H/2,
			W, H,
			fill)
		return
	}

	// Escape for XML
	switch c {
	case "&":
		c = "&amp;"
	case ">":
		c = "&gt;"
	case "<":
		c = "&lt;"
	}

	// usual case

	// Text elements <text> get an inline Y-offset of +4 – visually necessary for Y-alignment
	// with Dots to left or right.
	// The value +4 is in theory font-specific, but for common fonts it corresponds to the offset
	// from the center of the 8x16 cell and the "baseline" typical of Roman fonts, which
	// aligns with for example the bottom of the "bowl" of a lower-case 'g'.
	//     https://svgwg.org/svg2-draft/text.html#FontsGlyphs
	centerToBaseline := 4
	internal.MustFPrintf(out, `    <text x="%d" y="%d">%s</text>
`,
		p.X, p.Y+centerToBaseline, c)
}
