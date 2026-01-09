package svg

import (
	"fmt"
)

type (
	markArr [2]rune

	// "Binds" a <g> or <a> element, itself wrapping a series of <text> elements, to a
	// CSS class containing a property "goat-anchor-marks".
	//         https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Syntax/Introduction
	markBinding struct {

		goatAnchorClassStyleRuleProperties

 		// CSS class definitions -- one for each CSS style-rule with
		// property '"goat-anchor-substitutes"' matching 'markpair'
		ClassNames []string

		// Used only for CSS correctness-checking: should never appear in
		// selector list of a stylerule that contains GoAT-defined properties.
		// XX  ? No need to retain -- move out to a local var in the parsing loop?
		_idName string
	}

	// XX XX  Create a variant of this that identifies text ranges to be styled not by marks,
	//        but by match of the diagram text to a pattern (globbing, or regexp?)
	goatAnchorClassStyleRuleProperties struct {
		// For <text> elements, two characters long: Begin and end of a
		// left-to-right, top-to-bottom range within the input UTF-8 diagram
		// rectangle to which a markBinding should be applied.
		markpair markArr  // XX  unify with name of CSS property "goat-anchor-marks"    

		// Replacements for a markpair, if not zeroes, in which case no substitution is made.
		//subst markArr     // XX  unify with "goat-anchor-substitutes"   

		// If non-nil, an outgoing link to be emitted -- cannot be gotten indirectly
		// through a CSS class.
		//
		// The 'href' attribute will be bound to an enclosing Text "<a>" element
		// X Cannot be used in GitHub README.md Markdown files.
		HRef string       // XX  unify with "goat-anchor-href"   
	}

	// Reason for indirection to the referent "*markBinding": Other maps,
	// instantiated later, will want to refer to the same struct objects.
	MarkBindingMap map[markArr]*markBinding
)

const (
	goat_anchor_marks	  = "goat-anchor-marks"

	// XX  Need at least one example of beneficial use.
	//goat_anchor_substitutes	  = "goat-anchor-substitutes"

	goat_anchor_href  = "goat-anchor-href"
)

func (mb markBinding) String() string {
	return fmt.Sprintf(`
    markpair: %q` +
		//    `subst: %q` +
		`
    HRef: %q
    ClassNames: %q
    _idName: %q
`,
		mb.markpair.String(),
		//mb.subst.String(),
		mb.HRef,
		mb.ClassNames,
		mb._idName,
	)
}

var zeroMarkArr markArr = markArr{}
