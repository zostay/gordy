package gordy

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
)

// ATag is the type used to tag matches by type.
type ATag int

// A few standard tags for matches.
const (
	TNone ATag = iota
	TLiteral
	TLast ATag = 100_000
)

// Matcher is the type for matching functions. These accept a list of bytes.
// They return four values.
//
// The first return value is a *Match object. If this is returned, then it should return the number

// Matcher is the type for matching functions. These accept a list of bytes to
// start matching from and return a pointer to a Match and a list of remaining
// unmatched bytes.
//
// If the match is successful, then a pointer to a Match is returned and the
// remaining input is also returned. It is possible for a match to match zero
// bytes.
//
// If the match fails, then the Match should be returned as nil. Usually the
// remaining input is also returned as nil in that case.
type Matcher func(r *Parser) (*Match, error)

// BytePredicate is a function that returns true if it matches a single byte or
// false if it does not.
type BytePredicate func(c byte) bool

// RunePredicate is a function that returns true if it matches a single rune or
// false if it does not.
type RunePredicate func(r rune) bool

// Tracer is a function that is use to log or report parser traces. This
// function signature was chosen because it is commonly available, such as
// fmt.Print or log.Println, etc.
type Tracer func(v ...any)

const (
	stageTry = iota
	stageGot
	stageFail
)

type Parser struct {
	Trace Tracer

	parent *Parser
	buf    *Buffer
	r      *Reader
}

// New creates a new parser helper for recursive descent parsing using the
// default buffer size (inherited from bufio.Reader).
func New(r io.Reader) *Parser {
	buf := NewBuffer(r)
	return &Parser{
		buf: buf,
		r:   buf.Reader(),
	}
}

// NewSize creates a new parser helper for recursive descent parsing, but with a
// custom internal Buffer size.
func NewSize(r io.Reader, size int) *Parser {
	buf := NewBufferSize(r, size)
	return &Parser{
		buf: NewBufferSize(r, size),
		r:   buf.Reader(),
	}
}

func (p *Parser) trace(stage int, name string, args ...any) {
	if p.Trace != nil {
		out := &strings.Builder{}
		switch stage {
		case stageTry:
			fmt.Fprint(out, "TRY ")
		case stageGot:
			fmt.Fprint(out, "GOT ")
		}

		fmt.Fprint(out, name)
		fmt.Fprint(out, "(")

		var bs [10]byte
		n, _ := p.buf.peek(p.r.n, bs[:])
		fmt.Fprint(out, string(bs[:n]))
		fmt.Fprint(out, "…")

		for i, arg := range args {
			fmt.Fprint(out, ", ")

			if reflect.TypeOf(arg).Kind() == reflect.Func {
				fmt.Fprint(out, runtime.FuncForPC(reflect.ValueOf(arg).Pointer()).Name())
				continue
			}

			if i == len(args)-1 {
				if err, isErr := arg.(error); isErr {
					fmt.Fprintf(out, "): %v", err)
					return
				}

				if m, isMatch := arg.(*Match); isMatch {
					fmt.Fprintf(out, ") = %v", m)
					return
				}
			}

			fmt.Fprint(out, arg)
		}

		fmt.Print(")")
	}
}

// Read reads the next bytes from input.
func (p *Parser) Read(bs []byte) (int, error) {
	return p.r.Read(bs)
}

// ReadRunes reads teh next runes from input.
func (p *Parser) ReadRunes(rs []rune) (int, error) {
	return p.r.ReadRunes(rs)
}

// MayFail returns a new Parser that can be used to read input starting at the
// offset of the current Parser. Reads on the returned Parser will not impact
// the parent. When finished, you may call Keep on the child parser if you are
// ready to keep the reads made.
func (p *Parser) MayFail() *Parser {
	return &Parser{
		parent: p,
		buf:    p.buf,
		r:      p.r.Clone(),
	}
}

// Keep returns the parent Parser after updating it to have the same state as
// the child.
//
// When Keep is called on the root Parser object or its direct descendants, it
// will also free up memory by discarding data that won't be read again at the
// start of the buffer.
func (p *Parser) Keep() *Parser {
	// detect root or child of root cases
	var root *Parser
	if p.parent == nil {
		root = p
	} else if p.parent.parent == nil {
		root = p.parent
	}

	// when we are at or child of root, we can discard the read bytes
	if root != nil {
		root.buf.Collect(p.r)
		root.r.Reset()
		return root
	}

	// otherwise, we just want to make sure the parent moves forward to the
	// cursor position in the input so far
	p.parent.r = p.r
	return p.parent
}

// Discard returns the parent Parser without updating the state of the parent ot
// match the child.
func (p *Parser) Discard() *Parser {
	if p.parent != nil {
		return p.parent
	}
	return p
}

// MatchOneFunc matches exactly one byte if the next byte in the input matches
// the given predicate. If there's a match than a Match with the given ATag is
// returned for the matching byte. Otherwise, nil is returned.
func (p *Parser) MatchOneFunc(
	t ATag,
	pred BytePredicate,
) (*Match, error) {
	p = p.MayFail()

	var cs [1]byte
	_, err := p.Read(cs[:])
	if err != nil {
		p.trace(stageFail, "MatchOneFunc", t, pred, err)
		return nil, err
	}

	p.trace(stageTry, "MatchOneFunc", t, pred)

	if pred(cs[0]) {
		m := &Match{Tag: t, Content: cs[:]}
		p.trace(stageGot, "MatchOneFunc", t, pred, m)
		p = p.Keep()
		return m, nil
	}

	return nil, nil
}

// MatchOneFunc is a Matcher that calls the MatchOneFunc method.
func MatchOneFunc(
	t ATag,
	pred BytePredicate,
) Matcher {
	return func(p *Parser) (*Match, error) {
		return p.MatchOneFunc(t, pred)
	}
}

// MatchOne matches the next byte against the given byte values. If there's a
// match to any one of the values, it is returned.
func (p *Parser) MatchOne(
	t ATag,
	cs ...byte,
) (*Match, error) {
	p.trace(stageTry, "MatchOne", t, cs)
	return p.MatchOneFunc(t, func(b byte) bool {
		for _, c := range cs {
			if b == c {
				return true
			}
		}
		return false
	})
}

// MatchOne is a Matcher that calls the MatchOne method.
func MatchOne(
	t ATag,
	cs ...byte,
) Matcher {
	return func(p *Parser) (*Match, error) {
		return p.MatchOne(t, cs...)
	}
}

// MatchOneRuneFunc matches the next complete rune against the given
// RunePredicate.
func (p *Parser) MatchOneRuneFunc(
	t ATag,
	pred RunePredicate,
) (*Match, error) {
	p = p.MayFail()

	var rs [1]rune
	_, err := p.ReadRunes(rs[:])
	if err != nil {
		p.trace(stageFail, "MatchOneRuneFunc", t, pred, err)
		return nil, err
	}

	p.trace(stageTry, "MatchOneRuneFunc", t, pred)

	if pred(rs[0]) {
		m := &Match{Tag: t, Content: []byte(string(rs[:]))}
		p.Keep()
		return m, nil
	}

	return nil, nil
}

// MatchOneRuneFunc is a Matcher that calls the MatchOneRuneFunc method.
func MatchOneRuneFunc(
	t ATag,
	pred RunePredicate,
) Matcher {
	return func(p *Parser) (*Match, error) {
		return p.MatchOneRuneFunc(t, pred)
	}
}

// MatchOneRune matches the next complete run against the given literal rune. If
// found, a Match is returned.
func (p *Parser) MatchOneRune(
	t ATag,
	rs ...rune,
) (*Match, error) {
	p.trace(stageTry, "MatchOneRune", t, rs)
	return p.MatchOneRuneFunc(t, func(c rune) bool {
		for _, r := range rs {
			if c == r {
				return true
			}
		}
		return false
	})
}

// selectLongest is an internal helper used to find the longest match out of a
// list of matches.
func selectLongest(ms []*Match) int {
	var ln int
	var lm *Match

	for n, m := range ms {
		if lm == nil || m.Length() > lm.Length() {
			ln = n
			lm = m
		}
	}

	return ln
}

// MatchLongest tries all the given matchers against the current input. It then
// returns whichever of these matches works to match the most input.
func (p *Parser) MatchLongest(ms ...Matcher) (*Match, error) {
	msm := make([]*Match, len(ms))
	msp := make([]*Parser, len(ms))

	for i, mp := range ms {
		p := p.MayFail()
		m, err := mp(p)
		if err != nil {
			return nil, err
		}

		msm[i] = m
		msp[i] = p
	}

	if w := selectLongest(msm); w != -1 {
		p.trace(stageGot, "MatchLongest", w, msm[w])
		msp[w].Keep()
		return msm[w], nil
	}

	return nil, nil
}

// MatchManyWithSep matches the given matcher against the input provided that
// the separator matcher matches in between. It returns a match containing those
// matches. If fewer than min matches are present, the match returns no match.
func (p *Parser) MatchManyWithSep(
	t ATag,
	min int,
	mtch Matcher,
	sep Matcher,
) (*Match, error) {
	mbs := make([]*Match, 0)
	ms := make([]*Match, 0)
	totalLen := 0

	p.trace(stageTry, "MatchManyWithSep", t, min, mtch, sep)

	p := p.MayFail()

	for {
		var pms [2]*Match
		if len(ms) > 0 {
			m, err := sep(p)
			if err != nil {
				p.trace(stageFail, "MatchManyWithSep", t, min, mtch, sep, err)
				return nil, err
			}

			if m != nil {
				pms[0] = m
			} else {
				break
			}
		}

		m, err := mtch(p)
		if err != nil {
			p.trace(stageFail, "MatchManyWithSep", t, min, mtch, sep, err)
			return nil, err
		}

		if m != nil {
			pms[1] = m

			if len(ms) > 0 {
				totalLen += len(pms[0].Content)
			}
			totalLen += len(pms[1].Content)

			mbs = append(mbs, m)
			if len(ms) > 0 {
				ms = append(ms, pms[0], pms[1])
			} else {
				ms = append(ms, pms[1])
			}

			continue
		}

		break
	}

	if len(mbs) < min {
		return nil, nil
	}

	content := make([]byte, 0, totalLen)
	for _, m := range ms {
		content = append(content, m.Content...)
	}

	m := &Match{
		Tag:      t,
		Content:  content,
		Group:    map[string]*Match{},
		Submatch: mbs,
	}

	p.trace(stageGot, "MatchManyWithSep", t, min, mtch, sep, m)
	p.Keep()
	return m, nil
}

// MatchMany matches the given matcher as many times as possible one after
// another on the input. If the number of matches is fewer than min, it returns
// nil.
func (p *Parser) MatchMany(
	t ATag,
	min int,
	mtch Matcher,
) (*Match, error) {
	content := make([]byte, 0)
	ms := make([]*Match, 0, min)

	p = p.MayFail()

	for {
		m, err := mtch(p)
		if err != nil {
			return nil, err
		}

		if m != nil {
			ms = append(ms, m)
			content = append(content, m.Content...)

			continue
		}

		break
	}

	if len(ms) < min {
		return nil, nil
	}

	m := &Match{
		Tag:      t,
		Content:  content,
		Group:    map[string]*Match{},
		Submatch: ms,
	}

	p.trace(stageGot, "MatchMany", t, min, mtch, m)
	p = p.Keep()
	return m, nil
}
