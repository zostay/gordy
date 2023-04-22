package match

import (
	"github.com/zostay/go-std/slices"

	"github.com/zostay/gordy"
	"github.com/zostay/gordy/parser"
	"github.com/zostay/gordy/token"
)

// RunePredicate is a function that returns true if it matches a single rune or
// false if it does not.
type RunePredicate func(r rune) bool

// RunesInSet creates a RunePredicate from the set of runes given.
func RunesInSet(cs ...rune) RunePredicate {
	return func(r rune) bool {
		for _, c := range cs {
			if c == r {
				return true
			}
		}
		return false
	}
}

// RunesInRange creates a RunePredicate that matches any rune in the given
// range. The match is inclusive so runes equal to either end point are also
// matched.
func RunesInRange(cs, ce rune) RunePredicate {
	return func(r rune) bool {
		return r >= cs && r <= ce
	}
}

// AnyRunes creates a combined RunePredicate that matches a byte that matches
// any of the given predicates.
func AnyRunes(preds ...RunePredicate) RunePredicate {
	switch len(preds) {
	case 0:
		return func(rune) bool { return false }
	case 1:
		return preds[0]
	default:
		return func(r rune) bool {
			for _, pred := range preds {
				if pred(r) {
					return true
				}
			}
			return false
		}
	}
}

// NotRunes creates a combined RunePredicate that matches a rune that does not
// match any of the given predicates.
func NotRunes(preds ...RunePredicate) RunePredicate {
	return func(r rune) bool {
		for _, pred := range preds {
			if pred(r) {
				return false
			}
		}
		return true
	}
}

// ThisButNotThatRunes creates a combined RunePredicate that matches a rune that
// matches the first predicate, but does not match the second predicate.
func ThisButNotThatRunes(this, that RunePredicate) RunePredicate {
	return func(r rune) bool {
		if this(r) {
			if that(r) {
				return false
			}
			return true
		}
		return false
	}
}

// Runes is the Matcher returned by OneRune. It provides a number of tools that
// allow this Matcher to be combined with other Runes Matchers.
type Runes struct {
	t        token.Tag
	from, to int
	pred     RunePredicate
}

// OneRune returns a matcher that matches the next complete rune against the
// given RunePredicates. If that rune matches any of the given predicates, a
// Match is returned with the given token.Tag. If no match is found, nil will
// be returned.
func OneRune(
	t token.Tag,
	preds ...RunePredicate,
) gordy.Matcher {
	return &Runes{
		t:    t,
		from: 1,
		to:   1,
		pred: AnyRunes(preds...),
	}
}

// NRunes returns a Runes Matcher that matches the next complete runes in the
// input. You can specify the range of matches expected inclusive. If the
// correct number of runes is matched, a Match with the given token.Tag is
// returned. Otherwise, nil is returned.
func NRunes(
	t token.Tag,
	from, to int,
	preds ...RunePredicate,
) gordy.Matcher {
	return &Runes{
		t:    t,
		from: from,
		to:   to,
		pred: AnyRunes(preds...),
	}
}

// Match returns a Match with the configured token.Tag if the next byte in the
// input matches the predicate. It returns nil otherwise.
func (r *Runes) Match(p *gordy.Parser) (*parser.Match, error) {
	rs := make([]rune, r.from, r.from+r.to)
	for i := 0; i < r.from; i++ {
		c, ok, err := r.matchOne(p)
		if err != nil {
			p.Trace(gordy.StageFail, "Runes.Match", r.t, r.from, r.to, r.pred, i, err)
			return nil, err
		}

		p.Trace(gordy.StageTry, "Runes.Match", r.t, r.from, r.to, r.pred, i)
		if !ok {
			return nil, nil
		}

		rs[i] = c
	}

	for i := r.from; i < r.to; i++ {
		c, ok, err := r.matchOne(p)
		if err != nil {
			p.Trace(gordy.StageFail, "Runes.Match", r.t, r.from, r.to, r.pred, i, err)
			return nil, err
		}

		p.Trace(gordy.StageTry, "Runes.Match", r.t, r.from, r.to, r.pred, i)
		if !ok {
			break
		}

		rs = append(rs, c)
	}

	m := &parser.Match{Tag: r.t, Content: []byte(string(rs))}
	p.Trace(gordy.StageGot, "Runes.Match", r.t, r.from, r.to, r.pred, m)
	return m, nil
}

// matchOne returns the matched rune and true or zero and false if no rune was
// matched.
func (r *Runes) matchOne(p *gordy.Parser) (rune, bool, error) {
	var rs [1]rune
	_, err := p.ReadRunes(rs[:])
	if err != nil {
		return 0, false, err
	}

	if r.pred(rs[0]) {
		return rs[0], true, nil
	}

	return 0, false, nil
}

func extractPredFromRunes(r *Runes) RunePredicate {
	return r.pred
}

// AndAlso creates a new Runes Matcher which combines the predicate of this
// Runes Matcher with predicates of the given Runes Matchers such that a match
// occurs if the next byte in the input matches any of those predicates. The
// returned Match (when found), will have the token.Tag of this Runes Matcher.
func (r *Runes) AndAlso(rs ...*Runes) *Runes {
	preds := slices.Map(rs, extractPredFromRunes)
	slices.Unshift(preds, r.pred)
	return &Runes{
		t:    r.t,
		pred: AnyRunes(preds...),
	}
}

// ButNot creates a new Bytes Matcher which combines the predicate of this
// Bytes Matcher with predicates of the given Bytes Matchers such that a match
// is successful if it matches this Bytes Matcher, but not those.
func (r *Runes) ButNot(rs ...*Runes) *Runes {
	preds := slices.Map(rs, extractPredFromRunes)
	return &Runes{
		t:    r.t,
		pred: ThisButNotThatRunes(r.pred, AnyRunes(preds...)),
	}
}
