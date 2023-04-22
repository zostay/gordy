v0.1.0  2023-04-21

 * Completely rewrote the system.
 * Added the parser.Input for allowing parsing against an io.Reader that can be 
   optimized for memory efficiency as opposed to using slices, which are harder 
   to optimize.
 * Added parser.Buffer and parser.Reader to handle the low-level aspects of the 
   interface.
 * Moved gordy.ATag and related pieces to token.Tag
 * Added token.NextTag to allow for the creation of token.Tags that generate tag 
   serial numbers at runtime to avoid conflicts in cases where different parsers
   might be mixed together.
 * Renamed gordy.TNone to token.None and defined it's meaning more carefully.
 * Renamed gordy.TLiteral to token.Literal and defined it's meaning more 
   carefully.
 * Renamed gordy.TLast to token.Last and explained it's use more carefully.
 * Moved gordy.Matcher to parser.Matcher and converted this to an interface.
 * Added parser.MatcherFunc to define a parser.Matcher using the previous 
   interface, when a functor is appropriate.
 * Moved gordy.Match to parser.Match
 * Created match.Bytes for matching zero or more bytes.
 * Created match.Runes for matching zero or more runes.
 * Added an example.
 * Converted all the code that was a Matcher before into a Matcher generator.
 * Renamed gordy.LongestMatch to match.Longest
 * Renamed gordy.MatchManyWithSep to match.ManyWithSep
 * Renamed gordy.MatchMany to match.Many
 * Added match.First matcher as a short-circuiting matcher.
 * Added match.Seq to match a sequence of matchers.
 * Added match.SeqNamed to match a sequence of machers, with named values in the 
   resulting Match object.
 * Added match.Optional.
 * Added match.TryAndKeep.
 * Replaced gordy.MatchOne with match.OneByte and a set of special predicate
   generators named match.BytesInSet, match.BytesInRange, match.AnyBytes,
   match.NotBytes, match.ThisButNotThatBytes
 * Added match.NBytes to be used with the byte predicate generators to match
   multi-byte sequences.
 * Replaced gordy.MatchOneRune with match.OneRune and a set of special predicate
   generates named match.RunesInSet, match.RunesInRange, match.AnyRunes,
   match.NotRunes, match.ThisButNotThatRunes
 * Added match.NRunes to be used with the byte predicate generates to match 
   multi-rune sequeneces.
 * Exposed the tracing mechanism via the Trace method on parser.Input.

v0.0.0  2023-04-18

 * Initial release.
 * Was copied from a library built into another project, zostay/go-addr.
