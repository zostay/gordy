package token

// Tag is the abstract tag identifier used to tag matches by type in the
// constructed abstract syntax tree.
type Tag int

// A few standard tags for matches.
const (
	None Tag = iota
	Literal
	Last
)

var prevTag Tag = Last

func NextTag() Tag {
	prevTag++
	return prevTag
}
