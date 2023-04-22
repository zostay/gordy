package match

import (
	"github.com/zostay/gordy"
	"github.com/zostay/gordy/parser"
	"github.com/zostay/gordy/token"
)

// selectLongest is an internal helper used to find the longest match out of a
// list of matches.
func selectLongest(ms []*parser.Match) int {
	var ln int
	var lm *parser.Match

	for n, m := range ms {
		if lm == nil || m.Length() > lm.Length() {
			ln = n
			lm = m
		}
	}

	return ln
}

// Longest returns a Matcher that tries all the given matchers against the
// current input. It will keep the longest match found and discard the rest. It
// that longest Match.
func Longest(ms ...gordy.Matcher) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		msm := make([]*parser.Match, len(ms))
		msp := make([]*gordy.Parser, len(ms))

		for i, mp := range ms {
			p := p.MayFail()
			m, err := mp.Match(p)
			if err != nil {
				return nil, err
			}

			msm[i] = m
			msp[i] = p
		}

		if w := selectLongest(msm); w != -1 {
			p.Trace(gordy.StageGot, "MatchLongest", w, msm[w])
			msp[w].Keep()
			return msm[w], nil
		}

		return nil, nil
	}
}

// ManyWithSep returns a matcher that matches the given matcher against the
// input provided that the separator matcher matches in between. It returns a
// match containing those matches. If fewer than min matches are present, the
// match returns no match.
func ManyWithSep(
	t token.Tag,
	min int,
	mtch gordy.Matcher,
	sep gordy.Matcher,
) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		mbs := make([]*parser.Match, 0)
		ms := make([]*parser.Match, 0)
		totalLen := 0

		p.Trace(gordy.StageTry, "MatchManyWithSep", t, min, mtch, sep)

		for {
			var pms [2]*parser.Match
			if len(ms) > 0 {
				m, err := sep.Match(p)
				if err != nil {
					p.Trace(gordy.StageFail, "MatchManyWithSep", t, min, mtch, sep, err)
					return nil, err
				}

				if m != nil {
					pms[0] = m
				} else {
					break
				}
			}

			m, err := mtch.Match(p)
			if err != nil {
				p.Trace(gordy.StageFail, "MatchManyWithSep", t, min, mtch, sep, err)
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

		m := &parser.Match{
			Tag:      t,
			Content:  content,
			Group:    map[string]*parser.Match{},
			Submatch: mbs,
		}

		p.Trace(gordy.StageGot, "MatchManyWithSep", t, min, mtch, sep, m)
		return m, nil
	}
}

// Many returns a Matcher that matches the given matcher as many times as
// possible one after another on the input. If the number of matches is fewer
// than min, it returns nil.
func Many(
	t token.Tag,
	min int,
	mtch gordy.Matcher,
) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		content := make([]byte, 0)
		ms := make([]*parser.Match, 0, min)

		for {
			m, err := mtch.Match(p)
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

		m := &parser.Match{
			Tag:      t,
			Content:  content,
			Group:    map[string]*parser.Match{},
			Submatch: ms,
		}

		p.Trace(gordy.StageGot, "MatchMany", t, min, mtch, m)
		return m, nil
	}
}

// First returns a matcher that will try each match and immediately returns on
// the first one tried that succeeds. Returns no match if none succeed.
func First(mtchs ...gordy.Matcher) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		for _, mtch := range mtchs {
			p := p.MayFail()

			m, err := mtch.Match(p)
			if err != nil {
				return nil, err
			}

			if m != nil {
				return m, nil
			}
		}

		return nil, nil
	}
}

// Seq returns a Matcher that applies each passed Matcher in turn against the
// input. Returns with no match immediately if any Matcher in the sequence
// fails. Returns the whole Match if every Matcher succeeds.
func Seq(
	t token.Tag,
	mtchs ...gordy.Matcher,
) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		ms := make([]*parser.Match, len(mtchs))
		for i, mtch := range mtchs {
			m, err := mtch.Match(p)
			if err != nil || m == nil {
				return nil, err
			}

			ms[i] = m
		}

		return &parser.Match{
			Tag:      t,
			Submatch: ms,
		}, nil
	}
}

// SeqNamed returns a Matcher that applies each named Matcher in turn against
// the input. Returns with no match immediately if any Matcher in the sequence
// fails to match. Returns the whole Match if every Matcher succeeds. The
// Matchers passed must all have a name passed in the arguments. (If you want a
// submatch to be unnamed, pass the empty string.)
func SeqNamed(
	t token.Tag,
	ms ...any,
) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		mps := make([]any, len(ms))
		for i, mtch := range ms {
			if i%2 == 0 {
				continue
			}

			m, err := mtch.(gordy.Matcher).Match(p)
			if err != nil || m == nil {
				return nil, err
			}

			mps[i-1] = ms[i-1]
			mps[i] = m
		}

		return parser.BuildMatch(t, mps...), nil
	}
}

// Optional returns a Matcher that returns the Match when the called Matcher
// matches, but also returns an empty Match when the called Matcher does not
// match. The token.Tag on the empty Match is token.None.
func Optional(
	mtch gordy.Matcher,
) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		m, err := TryAndKeep(mtch).Match(p)
		if err != nil {
			return nil, err
		}

		if m != nil {
			return m, nil
		}

		return &parser.Match{Tag: token.None}, nil
	}
}

// TryAndKeep returns a matcher that will call the given Matcher and try to
// match against the input. On fail, input is restored to what it was before. On
// success, input moves forward to whatever the Matcher consumed.
func TryAndKeep(mtch gordy.Matcher) gordy.MatcherFunc {
	return func(p *gordy.Parser) (*parser.Match, error) {
		p = p.MayFail()

		m, err := mtch.Match(p)
		if err != nil {
			return nil, err
		}

		if m == nil {
			return nil, nil
		}

		p.Keep()
		return m, nil
	}
}
