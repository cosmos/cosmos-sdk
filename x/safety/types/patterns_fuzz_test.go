package types_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

// refPatternRE is an independent statement of the pattern grammar: "*", a
// prefix "/<ident>(.<ident>)*.*", or an exact type URL "/<ident>(.<ident>)*".
var refPatternRE = regexp.MustCompile(`^(\*|/[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*(\.\*)?)$`)

// refPattern is a parsed pattern used by the reference model.
type refPattern struct {
	wildcard bool
	prefix   bool
	segs     []string
}

// parseRef parses a pattern already accepted by refPatternRE.
func parseRef(p string) refPattern {
	if p == "*" {
		return refPattern{wildcard: true}
	}
	body := p[1:]
	prefix := strings.HasSuffix(body, ".*")
	body = strings.TrimSuffix(body, ".*")
	return refPattern{prefix: prefix, segs: strings.Split(body, ".")}
}

func segsHavePrefix(segs, prefix []string) bool {
	if len(prefix) > len(segs) {
		return false
	}
	for i := range prefix {
		if segs[i] != prefix[i] {
			return false
		}
	}
	return true
}

// refCovers is a segment-list definition of Covers, written independently of
// the string-prefix implementation.
func refCovers(s, p string) bool {
	rs, rp := parseRef(s), parseRef(p)
	switch {
	case rs.wildcard:
		return true
	case rp.wildcard:
		return false
	case s == p:
		return true
	case !rs.prefix:
		return false
	case rp.prefix:
		// "/a.*" covers "/a.b.*" (and itself, handled above).
		return len(rs.segs) <= len(rp.segs) && segsHavePrefix(rp.segs, rs.segs)
	default:
		// "/a.*" covers "/a.b" but not "/a".
		return len(rs.segs) < len(rp.segs) && segsHavePrefix(rp.segs, rs.segs)
	}
}

// refMatchesURL reports whether pattern p matches the concrete msg type URL.
func refMatchesURL(p, url string) bool { return refCovers(p, url) }

// concreteURL returns a msg type URL that pattern p matches.
func concreteURL(p string) string {
	switch {
	case p == "*":
		return "/any.Msg"
	case strings.HasSuffix(p, ".*"):
		return strings.TrimSuffix(p, "*") + "MsgZ"
	default:
		return p
	}
}

func refIsProtected(url string) bool {
	if url == "/cosmos.consensus.v1.MsgUpdateParams" {
		return true
	}
	for _, pre := range []string{"/cosmos.safety.", "/cosmos.gov.", "/cosmos.upgrade."} {
		if strings.HasPrefix(url, pre) {
			return true
		}
	}
	return false
}

func refIsWrapper(url string) bool {
	switch url {
	case "/cosmos.authz.v1beta1.MsgExec", "/cosmos.group.v1.MsgSubmitProposal",
		"/cosmos.group.v1.MsgVote", "/cosmos.group.v1.MsgExec":
		return true
	}
	return false
}

// refBlocks reports whether disabled entry d blocks pattern p: only an exact
// entry blocks a wrapper.
func refBlocks(d, p string) bool {
	if refIsWrapper(p) {
		return d == p
	}
	return refCovers(d, p)
}

// segAlphabet is small on purpose so that random patterns often share
// prefixes, and includes near-misses for the segment-boundary check.
var segAlphabet = []string{"a", "ab", "b", "cosmos", "bank", "banking", "v1", "MsgSend", "_x"}

// decodePattern builds a well-formed pattern from the front of data and
// returns the remaining bytes.
func decodePattern(data []byte) (string, []byte) {
	next := func() byte {
		if len(data) == 0 {
			return 0
		}
		b := data[0]
		data = data[1:]
		return b
	}
	head := next()
	if head%11 == 0 {
		return "*", data
	}
	n := 1 + int(head>>4)%4
	segs := make([]string, n)
	for i := range segs {
		segs[i] = segAlphabet[int(next())%len(segAlphabet)]
	}
	p := "/" + strings.Join(segs, ".")
	if head&1 == 1 {
		p += ".*"
	}
	return p, data
}

var patternSeeds = []string{
	"", "*", "**", "/*", "/.*", "/", "/a", "/a.*", "/a.b.*", "/a.b.c", "/cosmos.bank.v1beta1.MsgSend",
	"/cosmos.bank.*", "/cosmos.banking.MsgX", "/cosmos..bank.*", "/cosmos.bank.", "/cosmos.bank*",
	"/cosmos.*.MsgSend", "/cosmos.bank.*.*", "/1a.b", "/a-b.*", " /a", "/a ", "/_x.y_2.*", "/a.*\x00",
	"/é.*", "/a.\n", "*.*", "/a.**",
	"/" + strings.Repeat("a", 255), "/" + strings.Repeat("a", 256),
}

// FuzzValidatePattern checks ValidatePattern against the reference grammar
// and, for valid patterns, the single-pattern Covers properties.
func FuzzValidatePattern(f *testing.F) {
	for _, s := range patternSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, p string) {
		err := types.ValidatePattern(p)
		want := refPatternRE.MatchString(p) && len(p) <= 256
		if !want {
			require.ErrorIs(t, err, types.ErrInvalidPattern, "%q", p)
			require.Error(t, types.ValidatePatterns([]string{p}), "%q", p)
			return
		}
		require.NoError(t, err, "%q", p)
		require.NoError(t, types.ValidatePatterns([]string{p}))
		require.ErrorIs(t, types.ValidatePatterns([]string{p, p}), types.ErrDuplicatePattern)

		require.True(t, types.Covers(p, p), "reflexive: %q", p)
		require.True(t, types.Covers(types.Wildcard, p), "* covers %q", p)
		require.Equal(t, p == types.Wildcard, types.Covers(p, types.Wildcard), "only * covers *: %q", p)
		require.Equal(t, p != "*" && strings.HasSuffix(p, ".*"), types.IsPrefixPattern(p), "%q", p)
		require.True(t, refMatchesURL(p, concreteURL(p)))
	})
}

// checkCoversProperties asserts the algebraic properties of Covers on three
// well-formed patterns.
func checkCoversProperties(t *testing.T, a, b, c string) {
	t.Helper()
	for _, x := range []string{a, b, c} {
		require.True(t, types.Covers(x, x), "reflexive: %q", x)
		require.Equal(t, x == types.Wildcard, types.Covers(x, types.Wildcard), "only * covers *: %q", x)
	}
	pairs := [][2]string{{a, b}, {b, a}, {a, c}, {c, a}, {b, c}, {c, b}}
	for _, pr := range pairs {
		s, p := pr[0], pr[1]
		got := types.Covers(s, p)
		require.Equal(t, refCovers(s, p), got, "Covers(%q, %q)", s, p)
		blocks := types.Blocks(s, p)
		require.Equal(t, refBlocks(s, p), blocks, "Blocks(%q, %q)", s, p)
		if types.ValidateTypeURL(p) == nil {
			// BlockingPatterns lists exactly the disabled entries that block
			// the type URL p, so the keeper's point lookups see what Blocks
			// would.
			require.Equal(t, blocks, slices.Contains(slices.Collect(types.BlockingPatterns(p)), s), "BlockingPatterns(%q) and %q", p, s)
		}
		if got {
			// Soundness: whatever p matches, s matches too.
			require.True(t, refMatchesURL(s, concreteURL(p)), "Covers(%q, %q) but %q not matched", s, p, concreteURL(p))
			// Segment boundary: a prefix match always ends on a dot.
			if types.IsPrefixPattern(s) && s != p {
				stem := strings.TrimSuffix(s, "*")
				require.True(t, strings.HasPrefix(p, stem), "Covers(%q, %q)", s, p)
			}
			// Antisymmetry.
			if types.Covers(p, s) {
				require.Equal(t, s, p, "antisymmetry")
			}
		}
	}
	// Transitivity over every ordering.
	trip := [][3]string{{a, b, c}, {a, c, b}, {b, a, c}, {b, c, a}, {c, a, b}, {c, b, a}}
	for _, tr := range trip {
		x, y, z := tr[0], tr[1], tr[2]
		if types.Covers(x, y) && types.Covers(y, z) {
			require.True(t, types.Covers(x, z), "transitive: %q > %q > %q", x, y, z)
		}
	}
	require.Equal(t, types.Covers(a, c) || types.Covers(b, c), types.CoveredByAny([]string{a, b}, c))
	require.False(t, types.CoveredByAny(nil, a))
}

// FuzzCovers decodes three well-formed patterns over a small alphabet so
// that related patterns are common, then checks the Covers algebra.
func FuzzCovers(f *testing.F) {
	f.Add([]byte{0x01, 3, 0x12, 3, 4, 0x20, 3, 4, 7})
	f.Add([]byte{0x11, 3, 4, 0x20, 3, 5, 7, 0})
	f.Add([]byte{0x00, 0x13, 0, 1, 0x02, 0})
	f.Add([]byte{0x31, 0, 1, 2, 3, 0x20, 0, 1, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		a, data := decodePattern(data)
		b, data := decodePattern(data)
		c, _ := decodePattern(data)
		for _, x := range []string{a, b, c} {
			require.NoError(t, types.ValidatePattern(x), "decoder produced %q", x)
		}
		checkCoversProperties(t, a, b, c)
	})
}

// FuzzCoversStrings checks the Covers algebra on arbitrary strings that the
// grammar accepts, covering identifiers outside FuzzCovers' alphabet.
func FuzzCoversStrings(f *testing.F) {
	f.Add("/cosmos.bank.*", "/cosmos.bank.v1beta1.MsgSend", "/cosmos.banking.MsgX")
	f.Add("*", "/a.*", "/a.b.*")
	f.Add("/a.*", "/a", "/ab.c")
	f.Add("/a.b.*", "/a.*", "*")
	f.Add("/cosmos.authz.*", "/cosmos.authz.v1beta1.MsgExec", "*")
	f.Add("/cosmos.group.v1.*", "/cosmos.group.v1.MsgVote", "/cosmos.group.v1.MsgCreateGroup")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		for _, x := range []string{a, b, c} {
			if types.ValidatePattern(x) != nil {
				t.Skip()
			}
		}
		checkCoversProperties(t, a, b, c)
	})
}

// FuzzTypeURL checks IsProtected and PrefixPatterns on msg type URLs.
func FuzzTypeURL(f *testing.F) {
	for _, s := range []string{
		"/cosmos.bank.v1beta1.MsgSend", "/cosmos.gov.v1.MsgVote", "/cosmos.gov", "/cosmos.govx.MsgY",
		"/cosmos.safety.v1.MsgReset", "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		"/cosmos.consensus.v1.MsgUpdateParams", "/cosmos.consensus.v1.MsgUpdateParamsX", "cosmos.gov.v1.MsgVote",
		"/A", "/a.b", "/ibc.cosmos.gov.v1.MsgX",
		"/cosmos.authz.v1beta1.MsgExec", "/cosmos.authz.v1beta1.MsgExecX", "/cosmos.group.v1.MsgVote",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, url string) {
		require.Equal(t, refIsProtected(url), types.IsProtected(url), "%q", url)
		require.Equal(t, refIsWrapper(url), types.IsWrapper(url), "%q", url)
		require.False(t, types.IsWrapper(url) && types.IsProtected(url), "%q is both a wrapper and protected", url)

		if types.ValidatePattern(url) != nil || types.IsPrefixPattern(url) || url == types.Wildcard {
			return
		}
		// url is a well-formed exact type URL.
		got := slices.Collect(types.PrefixPatterns(url))
		require.Len(t, got, strings.Count(url, "."), "%q", url)
		for i, pre := range got {
			require.NoError(t, types.ValidatePattern(pre), "%q", pre)
			require.True(t, types.IsPrefixPattern(pre), "%q", pre)
			require.True(t, types.Covers(pre, url), "%q should cover %q", pre, url)
			if i > 0 {
				require.True(t, types.Covers(got[i-1], pre), "shortest first: %q, %q", got[i-1], pre)
			}
		}
		// Every prefix pattern covering url is listed: the router's point
		// lookups in IsAllowed see exactly what Covers would.
		segs := strings.Split(url[1:], ".")
		for n := 1; n < len(segs); n++ {
			pre := "/" + strings.Join(segs[:n], ".") + ".*"
			require.Contains(t, got, pre)
		}
		require.True(t, refMatchesURL(url, url))
	})
}
