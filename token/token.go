package token

// Tag is the abstract tag identifier used to tag matches by type in the
// constructed abstract syntax tree.
type Tag int

// A few standard tags for matches.
const (
	// None is the tag to use for matches that aren't actual matches, such as
	// are returned by match.Optional.
	None Tag = iota

	// Literal is the mot generic tag.
	Literal

	// Last identifies the first non-built-in tag. No guarantee is made that
	// this will never change.
	Last
)

var prevTag = Last

// NextTag provides an interface for assigning tags serial numbers at runtime to
// avoid conflicts between tags when parsers from different modules are mixed
// and matched. This returns the next available tag and should be called during
// init.
func NextTag() Tag {
	prevTag++
	return prevTag
}
