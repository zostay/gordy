package match

import (
	"github.com/zostay/go-std/slices"

	"github.com/zostay/gordy"
	"github.com/zostay/gordy/parser"
	"github.com/zostay/gordy/token"
)

// BytePredicate is a function that returns true if it matches a single byte or
// false if it does not.
type BytePredicate func(c byte) bool

// BytesInSet creates a BytePredicate from the set of bytes given.
func BytesInSet(cs ...byte) BytePredicate {
	return func(b byte) bool {
		for _, c := range cs {
			if c == b {
				return true
			}
		}
		return false
	}
}

// BytesInRange creates a BytePredicate that matches any byte in the given
// range. The match is inclusive so bytes equal to either end point are also
// matched.
func BytesInRange(cs, ce byte) BytePredicate {
	return func(b byte) bool {
		return b >= cs && b <= ce
	}
}

// AnyBytes creates a combined BytePredicate that matches a byte that matches
// any of the given predicates.
func AnyBytes(preds ...BytePredicate) BytePredicate {
	switch len(preds) {
	case 0:
		return func(byte) bool { return false }
	case 1:
		return preds[0]
	default:
		return func(b byte) bool {
			for _, pred := range preds {
				if pred(b) {
					return true
				}
			}
			return false
		}
	}
}

// NotBytes creates a combined BytePredicate that matches a byte that does not
// match any of the given predicates.
func NotBytes(preds ...BytePredicate) BytePredicate {
	return func(b byte) bool {
		for _, pred := range preds {
			if pred(b) {
				return false
			}
		}
		return true
	}
}

// ThisButNotThatBytes creates a combined BytePredicate that matches a byte that
// matches the first predicate, but does not match the second predicate.
func ThisButNotThatBytes(this, that BytePredicate) BytePredicate {
	return func(b byte) bool {
		if this(b) {
			if that(b) {
				return false
			}
			return true
		}
		return false
	}
}

// Bytes is the Matcher returned by OneByte. It provides a number of tools that
// allow this Matcher to be combined with other Bytes Matchers.
type Bytes struct {
	t        token.Tag
	from, to int
	pred     BytePredicate
}

// OneByte returns a Matcher that matches exactly one byte if the next byte in
// the input matches any of the given predicates. If there's a match then a
// Match with the given token.Tag is returned for the matching byte. Otherwise,
// nil is returned.
func OneByte(
	t token.Tag,
	preds ...BytePredicate,
) gordy.Matcher {
	return &Bytes{
		t:    t,
		pred: AnyBytes(preds...),
	}
}

// NBytes returns a Bytes Matcher that matches the next complete bytes in the
// input. You can specify the range of matches expected inclusive. If the
// correct number of bytes is matched, a Match with the given token.Tag is
// returned. Otherwise, nil is returned.
func NBytes(
	t token.Tag,
	from, to int,
	preds ...BytePredicate,
) gordy.Matcher {
	return &Bytes{
		t:    t,
		from: from,
		to:   to,
		pred: AnyBytes(preds...),
	}
}

// Match returns a Match with the configured token.Tag if the next byte in the
// input matches the predicate. It returns nil otherwise.
func (b *Bytes) Match(p *gordy.Parser) (*parser.Match, error) {
	bs := make([]byte, b.from, b.from+b.to)
	for i := 0; i <= b.from; i++ {
		c, ok, err := b.matchOne(p)
		if err != nil {
			p.Trace(gordy.StageFail, "Bytes.Match", b.t, b.from, b.to, b.pred, i, err)
			return nil, err
		}

		p.Trace(gordy.StageTry, "Bytes.Match", b.t, b.from, b.to, b.pred, i)
		if !ok {
			return nil, nil
		}

		bs[i] = c
	}

	for i := b.from + 1; i <= b.to; i++ {
		c, ok, err := b.matchOne(p)
		if err != nil {
			p.Trace(gordy.StageFail, "Bytes.Match", b.t, b.from, b.to, b.pred, i, err)
			return nil, err
		}

		p.Trace(gordy.StageTry, "Bytes.Match", b.t, b.from, b.to, b.pred, i)
		if !ok {
			break
		}

		bs = append(bs, c)
	}

	m := &parser.Match{Tag: b.t, Content: []byte(string(bs))}
	p.Trace(gordy.StageGot, "Bytes.Match", b.t, b.from, b.to, b.pred, m)
	return m, nil
}

// matchOne returns the matched byte and true or zero and false if no byte was
// matched.
func (b *Bytes) matchOne(p *gordy.Parser) (byte, bool, error) {
	var bs [1]byte
	_, err := p.Read(bs[:])
	if err != nil {
		return 0, false, err
	}

	if b.pred(bs[0]) {
		return bs[0], true, nil
	}

	return 0, false, nil
}

func extractPredFromBytes(b *Bytes) BytePredicate {
	return b.pred
}

// AndAlso creates a new Bytes Matcher which combines the predicate of this
// Bytes Matcher with predicates of the given Bytes Matchers such that a match
// occurs if the next byte in the input matches any of those predicates. The
// returned Match (when found), will have the token.Tag of this Bytes Matcher.
func (b *Bytes) AndAlso(bs ...*Bytes) *Bytes {
	preds := slices.Map(bs, extractPredFromBytes)
	slices.Unshift(preds, b.pred)
	return &Bytes{
		t:    b.t,
		pred: AnyBytes(preds...),
	}
}

// ButNot creates a new Bytes Matcher which combines the predicate of this
// Bytes Matcher with predicates of the given Bytes Matchers such that a match
// is successful if it matches this Bytes Matcher, but not those.
func (b *Bytes) ButNot(bs ...*Bytes) *Bytes {
	preds := slices.Map(bs, extractPredFromBytes)
	return &Bytes{
		t:    b.t,
		pred: ThisButNotThatBytes(b.pred, AnyBytes(preds...)),
	}
}
