package parser

// Matcher is the type for matching functions. These accept a list of bytes.
// They return four values.
//
// The first return value is a *Match object. If this is returned, then it should return the number
type Matcher interface {
	Match(p *Input) (*Match, error)
}

// MatcherFunc is the type for matching functions. These accept a list of bytes to
// start matching from and return a pointer to a Match and a list of remaining
// unmatched bytes.
//
// If the match is successful, then a pointer to a Match is returned and the
// remaining input is also returned. It is possible for a match to match zero
// bytes.
//
// If the match fails, then the Match should be returned as nil. Usually the
// remaining input is also returned as nil in that case.
type MatcherFunc func(r *Input) (*Match, error)

func (mfun MatcherFunc) Match(p *Input) (*Match, error) {
	return mfun(p)
}
