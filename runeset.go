package goat

type (
	empty [0]byte
	RuneSet map[rune]empty
)

// queries
func (set RuneSet) Contains(r rune) bool {
	_, found := set[r]
	return found
}
func (rs RuneSet) Slice() (s []rune) {
	for r := range rs {
		s = append(s, r)
	}
	return
}

// constructors
func MakeRuneSet(runes ...rune) RuneSet {
	rs := make(RuneSet)
	rs.ExtendSet(runes...)
	return rs
}
func (rs RuneSet) UnionSet(s RuneSet) {
	for r := range s {
		rs[r] = empty{}
	}
	return
}
func CopySet(s RuneSet) RuneSet {
	rs := make(RuneSet)
	rs.UnionSet(s)
	return rs
}
func (rs RuneSet) ExtendSet(runes ...rune) {
	for _, r := range runes {
		rs[r] = empty{}
	}
	return
}
func UnionSets(rss ...RuneSet) RuneSet {
	union := make(RuneSet)
	for _, rs := range rss {
		for r := range rs {
			union[r] = empty{}
		}
	}
	return union
}
