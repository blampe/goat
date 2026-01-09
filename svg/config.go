package svg

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"

	"github.com/blampe/goat"
	"github.com/blampe/goat/internal"
)

type (
	Config struct {
		LineFilter *regexp.Regexp

		beginMap,
		endMap map[rune]*markBinding
	}
)

const DUMP_CSS = false  // XX  Also suppresses some early panic calls. Bad?
func printfStderr(format string, args ...any) {
	fmt.Fprintf(os.Stderr,
		//"%s: "+format,   XX  why is this wrong?
		"%24s: %v\n",
		internal.Where(2), args)
		//    "%24s %16s() %v\n",   XX Who(2) always fails -- why?
		//    internal.Where(2), internal.Who(2), args)
}

// Extend MarkBindingMap according to any GoAT-specific properties found in 'cssBytes'.
func ParseCss(bindings MarkBindingMap, cssBytes []byte) error {
	parser := css.NewParser(parse.NewInputBytes(cssBytes), false)

	// Running accumulation of contents of current CSS RuleSet.
	// To be appended to returned slice 'bindings' iff at least one
	// GoAT pseudo-CSS-property is found.
	var markBinding markBinding

	// Iterate over the top level of the CSS structure.
	for tokenCount := 0; ; tokenCount++ {
		// X  Conform to CSS convention: GoAT-specific properties are read
		//    case-insensitively from CSS files
		grammarType, tokenType, tokenData := parser.Next()
		_ = tokenType

		switch grammarType {
		case css.ErrorGrammar:
			err := parser.Err()
			if err == io.EOF {
				return nil
			}
			// X  Interleave error descriptions with echoing of normal output.
			errData := []byte("ERROR(")
			for _, val := range parser.Values() {
				errData = append(errData, val.Data...)
			}
			errData = append(errData, ")"...)
			return fmt.Errorf("%v: %v %s\n", err, grammarType, string(errData))
		}

		if DUMP_CSS {
//			if tokenCount == 0 {
//				printfStderr(`%16v %16v %v
//`,				"grammarType", "tokenType", "string(tokenData)")
//			}
			printfStderr(`%#v %#v "%#v"
`,
				grammarType, tokenType, string(tokenData))
		}

		dumpValues := func(grammarType css.GrammarType) {
			if ! DUMP_CSS {
				return
			}
			printfStderr(
`                     %s Values():
`,
				grammarType.String())
		}

		switch grammarType {
		case css.BeginRulesetGrammar:
			dumpValues(grammarType)
			//  Enter a sub-loop at this point, for any CSS .class or #id block
			//  that requires examination.
			markBinding = beginRuleSet(parser)
		case css.EndRulesetGrammar:
			if markBinding.markpair != zeroMarkArr {
				if len(markBinding._idName) > 0 {
					return fmt.Errorf("Ruleset containing %q " +
						"was specified by an element tag (%q) -- NOT supported.",
						goat_anchor_marks, markBinding._idName)
				}

				mustInsertBinding(bindings, markBinding)
			} else {
				if len(markBinding.HRef) > 0 {
					return fmt.Errorf("Ruleset for classes %v contained %q, but no %q.",
						markBinding.ClassNames, goat_anchor_href, goat_anchor_marks)
				}
				// otherwise discard the markBinding
			}
		case css.DeclarationGrammar:
			// a single declaration within some CSS declaration block
			//   https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Syntax/Introduction#css_rulesets
			dumpValues(grammarType)
			cssTokens := parser.Values()
			// 'tokenData' here is the CSS property name
			switch string(tokenData) {
			default:
				continue   // assumed to be an ordinary property
			case goat_anchor_marks:
				markStr := lexDeclaration(cssTokens)
				markBinding.markpair = validPair(markStr)
			//case goat_anchor_substitutes:
			//	substStr := lexDeclaration(cssTokens)
			//	markBinding.subst = validPair(substStr)
			case goat_anchor_href:
				// XX  ? Feature wanted: Rebase local links from directory of TXT source to
				//     that of the SVG output?   
				markBinding.HRef = lexDeclaration(cssTokens)
			}
			// Fatal error if a GoAT-specific property is found
			// inside a RuleSet whose CSS selector list is not names of classes, with or without a prepended element tag.
			if len(markBinding.ClassNames) == 0 {
				// XX  At this point, markBinding.markpair[2] is not yet initialized.
				return fmt.Errorf(
					`GoAT-specific properties found, but no .CLASSNAME selectors for the RuleSet:
	%s`,
					markBinding.String())
			}
		}
	}
}

func validPair(str string) (_ markArr) {
	runes := []rune(str)  // string to slice conversion
	switch len(runes) {
	case 0:
		return
	case 2:
		// usual case
		return markArr(runes)
	default:
		log.Fatalf(
			"invalid mark pair length): %d", len(runes))
	}
	return
}

// Most declarations will be parsed into only one value; an example of
// the exceptional case would be "var(--red)"
func lexDeclaration(cssTokens []css.Token) (tokenStr string) {
	if len(cssTokens) > 1 {
		log.Fatalln("could this case arise???")   
	}
	tokenStr = strings.Trim(string(cssTokens[0].Data), "\"")
	if DUMP_CSS {
		printfStderr(
				`                         %16v: %s
`,
			cssTokens[0].TokenType, tokenStr)
	}
	return
}

func beginRuleSet(parser *css.Parser) (kb markBinding) {
	var isClass bool

	// Consume the CSS selector list.
	//   https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/Selectors/Selector_list
	// lastTokenStr := ""
	for _, token := range parser.Values() {
		tokenStr := string(token.Data)
		if DUMP_CSS {
			printfStderr(
				`                         %16v: %s
`,
				token.TokenType, tokenStr)
		}
		switch token.TokenType {
		case css.HashToken:
			// Introduces an HTML element ID string.
			// X The parser does not split out the initial '#' from the following identifier
			kb._idName = tokenStr[1:]
		case css.DelimToken:
			// The initial '.' of a class rule hits this.
			if tokenStr == "." {
				isClass = true
			}
 		case css.IdentToken: // X  hits for both a class name, and an SVG element tag, which is not needed.
			if isClass {
				kb.ClassNames = append(kb.ClassNames, tokenStr)
			}
		default:
			// some other selector syntax
		}
		//lastTokenStr = tokenStr
	}
	return
}

func (ma markArr) String() string {
	return fmt.Sprintf("%s%s", string(ma[0]), string(ma[1]))
}

// General model for merging content of CSS files is that CSS definitions are
// concatenated within the SVG -- no attempt to detect duplications.
func mustInsertBinding(bindings MarkBindingMap, kb markBinding) {
	seniorValue, exists := bindings[kb.markpair]
	if exists {
		if len(seniorValue.HRef) > 0 && len(kb.HRef) > 0 {
			log.Fatalf(`
Attempt to overwrite HRef:
    new: %+v
    old: %+v
`,
				seniorValue, kb)
		}
		seniorValue.HRef = kb.HRef

		// Allow colliding markBindings, with contents merged, following
		// existing practice of CSS rules.
		seniorValue.ClassNames = append(seniorValue.ClassNames, kb.ClassNames...)
		return
	}
	bindings[kb.markpair] = &kb  // escape to heap
}

func appendAttr(l, r string) string {
	if r == "" {
		return l
	} else if l == "" {
		return r
	} else {
		return l +`;
    ` + r
	}
}

func NewConfig(reservedSet goat.RuneSet,
	parsedCss MarkBindingMap,   // heavyweight object
) (Config, error) {

	// XX  Could be pushed into CSS parsing loop:
	//       ++ fewer loops, but ...
	//       -- parser would become sensitive to 'reservedSet'
	err := vetMarkBindingMap(reservedSet, parsedCss)
	if err != nil {
		return Config{}, err
	}

	conf := Config{
		// Copies of arg 'parsedCss', with possible edits.
		beginMap: make(map[rune]*markBinding),
		endMap:   make(map[rune]*markBinding),
	}
	for markpair, kb := range parsedCss {
		_ = markpair
		// Verify that kb.markpair[0] and kb.markpair[1] are both unique
		// across conf.MarkBindingMap, to avoid silently "orphaning" markBinding definitions.
		allocIfUnique := func(beMap map[rune]*markBinding, be rune) error {
			senior, nonUnique := beMap[be]
			if nonUnique {
				return fmt.Errorf("Mark already in use %s\nby markBinding:\n\t%+v",
					string(be), senior)
			}
			beMap[be] = kb
			return nil
		}
		// Afford the user identical begin and end marks.
		errBegin := allocIfUnique(conf.beginMap, kb.markpair[0])
		if errBegin != nil {
			return Config{}, errors.Join(errors.New("Begin "), errBegin)
		}
		errEnd   := allocIfUnique(conf.endMap, kb.markpair[1])
		if errEnd != nil {
			return Config{}, errors.Join(errors.New("End "), errEnd)
		}
	}
	return conf, nil
}

func vetMarkBindingMap(reservedSet goat.RuneSet, parsedCss MarkBindingMap) (err error) {
	for markRunes, kb := range parsedCss {
		if markRunes == zeroMarkArr {
			return errors.New(fmt.Sprintf(
				"invalid markpair (%s)", markRunes.String()))
		}
		err = vetMark(reservedSet, kb, markRunes[0])
		if err != nil {
			return err
		}
		err = vetMark(reservedSet, kb, markRunes[1])
		if err != nil {
			return err
		}
		kb.markpair = markArr(markRunes)
	}
	return
}

func vetMark(reservedSet goat.RuneSet, candidateMarkBinding *markBinding, r rune) error {
	_, found := reservedSet[r]
	if found {
		return errors.New(fmt.Sprintf(
			`reserved rune '%c' (0x%x) ` +
				`cannot be used as a markpairclass mark by markpairclass: %s`,
			r, r, formatMarkBinding(candidateMarkBinding)))
	}
	return nil
}
