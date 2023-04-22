package match_test

import (
	"fmt"
	"strings"

	"github.com/zostay/gordy"
	"github.com/zostay/gordy/match"
	"github.com/zostay/gordy/token"
)

func Example() {
	var (
		TDotAtom      = token.NextTag()
		TEmailAddress = token.NextTag()
		TAreaCode     = token.NextTag()
		TLocalCode    = token.NextTag()
		TPersonalCode = token.NextTag()
		TPhoneNumber  = token.NextTag()
	)

	var (
		MatchAlpha = match.OneByte(token.Literal,
			match.BytesInRange('a', 'z'),
			match.BytesInRange('A', 'Z'),
		)

		digits     = match.BytesInRange('0', '9')
		MatchDigit = match.OneByte(token.Literal, digits)

		MatchAText = match.First(
			MatchAlpha,
			MatchDigit,
			match.OneByte(token.Literal,
				match.BytesInSet(
					'!', '#', '$', '%', '&', '\'', '*', '+', '-', '/',
					'=', '?', '^', '_', '`', '{', '|', '}', '~',
				),
			),
		)

		MatchDotAtom = match.ManyWithSep(TDotAtom, 1,
			match.Many(token.Literal, 1, MatchAText),
			match.OneByte(token.Literal, match.BytesInSet('.')),
		)

		MatchLocalPart = MatchDotAtom
		MatchDomain    = MatchDotAtom

		MatchEmailAddress = match.SeqNamed(TEmailAddress,
			MatchLocalPart,
			match.OneByte(token.Literal, match.BytesInSet('@')),
			MatchDomain,
		)

		MatchAreaCode     = match.NBytes(TAreaCode, 3, 3, digits)
		MatchLocalCode    = match.NBytes(TLocalCode, 3, 3, digits)
		MatchPersonalCode = match.NBytes(TPersonalCode, 4, 4, digits)
		MatchHyphen       = match.OneByte(token.Literal, match.BytesInSet('-'))

		MatchPhoneNumber = match.Seq(
			TPhoneNumber,
			MatchAreaCode,
			match.Optional(MatchHyphen),
			MatchLocalCode,
			match.Optional(MatchHyphen),
			MatchPersonalCode,
		)

		MatchContactInfo = match.Longest(
			MatchPhoneNumber,
			MatchEmailAddress,
		)
	)

	contact := "555-555-5555"
	p := gordy.New(strings.NewReader(contact))
	m, err := MatchContactInfo.Match(p)
	if err != nil {
		panic(err)
	}

	fmt.Println(m)
}
