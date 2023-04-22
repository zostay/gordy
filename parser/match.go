package parser

import "github.com/zostay/gordy/token"

// Match is the object used to represent some segment of a parsed string.
type Match struct {
	Tag      token.Tag         // an identifier describing what the match represents
	Content  []byte            // the full content of the match
	Group    map[string]*Match // identifies named submatches
	Submatch []*Match          // identifies a list of submatches
	Made     interface{}       // a place to put high-level objects generated from this match
}

// Length returns the number of bytes matched for this match.
func (m *Match) Length() int {
	if m != nil {
		return len(m.Content)
	} else {
		return 0
	}
}

// BuildMatch is a short hand for building a match with named submatches.
func BuildMatch(t token.Tag, ms ...any) (m *Match) {
	g := make(map[string]*Match, len(ms)/2)
	s := make([]*Match, 0, len(ms)/2)
	c := make([]byte, 0)
	var n string
	for i, x := range ms {
		if i%2 == 0 {
			n = x.(string)
		} else if x.(*Match) != nil {
			if n != "" {
				g[n] = x.(*Match)
			}
			s = append(s, x.(*Match))
			c = append(c, x.(*Match).Content...)
		}
	}

	m = &Match{Tag: t, Content: c, Group: g, Submatch: s}
	// traceMatch("BuildMatch(%+v)", m)

	return
}
