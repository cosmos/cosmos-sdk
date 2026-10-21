package types

import (
	"iter"
	"slices"
	"strings"
)

// Wildcard is the pattern that covers every msg type URL.
const Wildcard = "*"

// prefixSuffix ends every prefix pattern.
const prefixSuffix = ".*"

// MaxPatternLength bounds the length of a pattern or msg type URL, well above
// any real proto full name.
const MaxPatternLength = 256

// protectedPrefixes are the proto package prefixes whose msgs can never be
// disabled. Keeping them executable guarantees the authority can always reset
// a trip, change params, run governance, and upgrade the chain.
var protectedPrefixes = []string{
	"cosmos.safety.",
	"cosmos.gov.",
	"cosmos.upgrade.",
}

// protectedExact are individual protected msg type URLs.
var protectedExact = []string{
	"/cosmos.consensus.v1.MsgUpdateParams",
}

// wrappers are the msg type URLs that dispatch other msgs through the
// router, which checks each inner msg. Group MsgSubmitProposal and MsgVote
// also record a proposal or a vote, and with EXEC_TRY they call the group
// msg server's Vote and Exec directly, which the msg server guard checks.
// An authority, operator or validator acting through authz or a group
// policy needs them to reach MsgReset and votes during an incident, so only
// an exact disabled entry blocks a wrapper; "*" and prefix entries do not.
// They are not protected: an exact trip still halts their handlers.
var wrappers = []string{
	"/cosmos.authz.v1beta1.MsgExec",
	"/cosmos.group.v1.MsgSubmitProposal",
	"/cosmos.group.v1.MsgVote",
	"/cosmos.group.v1.MsgExec",
}

// IsProtected reports whether a msg type URL belongs to the hardcoded
// protected set. IsAllowed always allows a protected type URL.
func IsProtected(typeURL string) bool {
	for _, exact := range protectedExact {
		if typeURL == exact {
			return true
		}
	}
	name, ok := strings.CutPrefix(typeURL, "/")
	if !ok {
		return false
	}
	for _, prefix := range protectedPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// IsWrapper reports whether a msg type URL belongs to the hardcoded wrapper
// set, which only an exact disabled entry blocks.
func IsWrapper(typeURL string) bool {
	return slices.Contains(wrappers, typeURL)
}

// IsPrefixPattern reports whether p has the prefix form "/a.b.*".
// It does not validate p.
func IsPrefixPattern(p string) bool {
	return p != Wildcard && strings.HasSuffix(p, prefixSuffix)
}

// ValidatePattern checks that p is "*", a prefix pattern "/<ident>(.<ident>)*.*",
// or an exact msg type URL "/<ident>(.<ident>)*".
func ValidatePattern(p string) error {
	if p == Wildcard {
		return nil
	}
	if len(p) > MaxPatternLength {
		return ErrInvalidPattern.Wrapf("%d bytes, limit %d", len(p), MaxPatternLength)
	}
	body, ok := strings.CutPrefix(p, "/")
	if !ok {
		return ErrInvalidPattern.Wrapf("%q: must be %q or start with \"/\"", p, Wildcard)
	}
	body = strings.TrimSuffix(body, prefixSuffix)
	if body == "" {
		return ErrInvalidPattern.Wrapf("%q: missing proto name", p)
	}
	for seg := range strings.SplitSeq(body, ".") {
		if !isProtoIdent(seg) {
			return ErrInvalidPattern.Wrapf("%q: %q is not a proto identifier", p, seg)
		}
	}
	return nil
}

// ValidateTypeURL checks that u is a well-formed exact msg type URL
// "/<ident>(.<ident>)*", not "*" or a prefix pattern.
func ValidateTypeURL(u string) error {
	if u == Wildcard || IsPrefixPattern(u) {
		return ErrInvalidPattern.Wrapf("%.40q: not a msg type URL", u)
	}
	return ValidatePattern(u)
}

// ValidatePatterns checks a non-empty list of well-formed, unique patterns.
func ValidatePatterns(patterns []string) error {
	if len(patterns) == 0 {
		return ErrEmptyPatterns
	}
	return validateUniquePatterns(patterns)
}

// validateUniquePatterns checks that every pattern is well-formed and unique.
// An empty list passes.
func validateUniquePatterns(patterns []string) error {
	seen := make(map[string]struct{}, len(patterns))
	for _, p := range patterns {
		if err := ValidatePattern(p); err != nil {
			return err
		}
		if _, dup := seen[p]; dup {
			return ErrDuplicatePattern.Wrap(p)
		}
		seen[p] = struct{}{}
	}
	return nil
}

// isProtoIdent reports whether s matches [A-Za-z_][A-Za-z0-9_]*.
func isProtoIdent(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '_', 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
		case '0' <= c && c <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

// Covers reports whether pattern s covers pattern p. "*" covers everything;
// any pattern covers itself; a prefix pattern "X.*" covers any exact or prefix
// pattern that starts with "X." at a segment boundary. Only "*" covers "*".
// Both arguments are assumed to be well-formed.
func Covers(s, p string) bool {
	if s == Wildcard || s == p {
		return true
	}
	if p == Wildcard || !IsPrefixPattern(s) {
		return false
	}
	// "/cosmos.bank.*" -> "/cosmos.bank."
	return strings.HasPrefix(p, strings.TrimSuffix(s, "*"))
}

// CoveredByAny reports whether some entry in scope covers p.
func CoveredByAny(scope []string, p string) bool {
	for _, s := range scope {
		if Covers(s, p) {
			return true
		}
	}
	return false
}

// Blocks reports whether the disabled entry d blocks what the well-formed
// pattern p matches: d covers p, and when p is a wrapper type URL, d is p
// itself. Scope checks use Covers, so a scope of "*" or "/cosmos.authz.*"
// still lets the operator trip and reset a wrapper by its exact URL.
func Blocks(d, p string) bool {
	return d == p || (!IsWrapper(p) && Covers(d, p))
}

// BlockingPatterns yields every disabled-list key that blocks the msg type
// URL typeURL: typeURL itself, then, unless typeURL is a wrapper, "*" and
// each segment-boundary prefix pattern of typeURL.
func BlockingPatterns(typeURL string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if !yield(typeURL) || IsWrapper(typeURL) || !yield(Wildcard) {
			return
		}
		for q := range PrefixPatterns(typeURL) {
			if !yield(q) {
				return
			}
		}
	}
}

// PrefixPatterns yields every segment-boundary prefix pattern of a msg type
// URL, shortest first. For "/cosmos.bank.v1beta1.MsgSend" it yields
// "/cosmos.*", "/cosmos.bank.*" and "/cosmos.bank.v1beta1.*".
func PrefixPatterns(typeURL string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for i := 1; i < len(typeURL); i++ {
			if typeURL[i] == '.' && !yield(typeURL[:i]+prefixSuffix) {
				return
			}
		}
	}
}
