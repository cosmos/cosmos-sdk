package types_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

func TestValidatePattern(t *testing.T) {
	valid := []string{
		"*",
		"/cosmos.*",
		"/cosmos.bank.*",
		"/cosmos.bank.v1beta1.*",
		"/cosmos.bank.v1beta1.MsgSend",
		"/ibc.applications.transfer.v1.MsgTransfer",
		"/_x.y_2.*",
		"/MsgTopLevel",
		"/a9.v1beta9.Msg9",
		"/AZaz_09.*",
	}
	for _, p := range valid {
		require.NoError(t, types.ValidatePattern(p), p)
	}

	invalid := []string{
		"",
		"**",
		"/*",
		"/.*",
		"/",
		"cosmos.bank.v1beta1.MsgSend",
		"/cosmos..bank.*",
		"/cosmos.bank.",
		"/cosmos.bank*",
		"/cosmos.*.MsgSend",
		"/cosmos.bank.v1beta1.MsgSend ",
		" /cosmos.bank.v1beta1.MsgSend",
		"/cosmos.1bank.*",
		"/cosmos.bank-v1.*",
		"/cosmos.bank.*.*",
		"/cosmos/bank.MsgSend",
	}
	for _, p := range invalid {
		require.ErrorIs(t, types.ValidatePattern(p), types.ErrInvalidPattern, "%q", p)
	}
}

func TestValidatePatterns(t *testing.T) {
	require.ErrorIs(t, types.ValidatePatterns(nil), types.ErrEmptyPatterns)
	require.ErrorIs(t, types.ValidatePatterns([]string{}), types.ErrEmptyPatterns)
	require.ErrorIs(t, types.ValidatePatterns([]string{"/a.*", "/a.*"}), types.ErrDuplicatePattern)
	require.ErrorIs(t, types.ValidatePatterns([]string{"/a.*", "bad"}), types.ErrInvalidPattern)
	require.NoError(t, types.ValidatePatterns([]string{"/a.*", "/a.b.*", "*"}))
}

func TestCovers(t *testing.T) {
	cases := []struct {
		s, p string
		want bool
	}{
		{"*", "*", true},
		{"*", "/cosmos.*", true},
		{"*", "/cosmos.bank.v1beta1.MsgSend", true},
		{"/cosmos.*", "*", false},
		{"/cosmos.bank.v1beta1.MsgSend", "*", false},
		{"/cosmos.*", "/cosmos.*", true},
		{"/cosmos.*", "/cosmos.bank.*", true},
		{"/cosmos.*", "/cosmos.bank.v1beta1.MsgSend", true},
		{"/cosmos.bank.*", "/cosmos.*", false},
		{"/cosmos.bank.*", "/cosmos.bank.v1beta1.MsgSend", true},
		// segment boundary: bank does not cover banking
		{"/cosmos.bank.*", "/cosmos.banking.MsgX", false},
		{"/cosmos.bank.*", "/cosmos.banking.*", false},
		{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend", true},
		{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSendX", false},
		{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.*", false},
		{"/cosmos.staking.*", "/cosmos.bank.v1beta1.MsgSend", false},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, types.Covers(tc.s, tc.p), "Covers(%q, %q)", tc.s, tc.p)
	}

	require.True(t, types.CoveredByAny([]string{"/x.*", "/cosmos.bank.*"}, "/cosmos.bank.v1beta1.MsgSend"))
	require.False(t, types.CoveredByAny(nil, "/cosmos.bank.v1beta1.MsgSend"))
}

func TestIsProtected(t *testing.T) {
	protected := []string{
		"/cosmos.safety.v1.MsgReset",
		"/cosmos.safety.v1.MsgTrip",
		"/cosmos.safety.v1.MsgUpdateParams",
		"/cosmos.gov.v1.MsgVote",
		"/cosmos.gov.v1beta1.MsgSubmitProposal",
		"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		"/cosmos.consensus.v1.MsgUpdateParams",
	}
	for _, u := range protected {
		require.True(t, types.IsProtected(u), u)
	}

	unprotected := []string{
		"/cosmos.bank.v1beta1.MsgSend",
		"/cosmos.consensus.v1.MsgOther",
		"/cosmos.governance.v1.MsgVote",
		"/cosmos.safetyx.v1.MsgTrip",
		"/cosmos.upgrades.v1.MsgX",
		"/cosmos.authz.v1beta1.MsgGrant",
		// wrappers are not protected, so an exact trip can halt their handlers
		"/cosmos.authz.v1beta1.MsgExec",
		"/cosmos.group.v1.MsgSubmitProposal",
		"/cosmos.group.v1.MsgVote",
		"/cosmos.group.v1.MsgExec",
		"/cosmos.group.v1.MsgCreateGroup",
		// protected package names later in the URL do not count
		"/ibc.cosmos.gov.v1.MsgX",
		"/other.cosmos.upgrade.v1.MsgX",
		"/x.cosmos.safety.v1.MsgReset",
		"cosmos.gov.v1.MsgVote",
		"*",
		"/cosmos.*",
	}
	for _, u := range unprotected {
		require.False(t, types.IsProtected(u), u)
	}
}

func TestIsWrapper(t *testing.T) {
	for _, u := range []string{
		"/cosmos.authz.v1beta1.MsgExec",
		"/cosmos.group.v1.MsgSubmitProposal",
		"/cosmos.group.v1.MsgVote",
		"/cosmos.group.v1.MsgExec",
	} {
		require.True(t, types.IsWrapper(u), u)
		require.False(t, types.IsProtected(u), u)
	}
	for _, u := range []string{
		"/cosmos.authz.v1beta1.MsgGrant",
		"/cosmos.authz.v1beta1.MsgExecX",
		"/cosmos.authz.v1.MsgExec",
		"/cosmos.group.v1.MsgCreateGroup",
		"/cosmos.group.v1.MsgVoteX",
		"/cosmos.gov.v1.MsgSubmitProposal",
		"/cosmos.gov.v1.MsgVote",
		"/cosmos.authz.*",
		"/cosmos.group.v1.*",
		"cosmos.authz.v1beta1.MsgExec",
		"*",
	} {
		require.False(t, types.IsWrapper(u), u)
	}
}

func TestBlocks(t *testing.T) {
	const exec = "/cosmos.authz.v1beta1.MsgExec"
	cases := []struct {
		d, p string
		want bool
	}{
		{"*", "/cosmos.bank.v1beta1.MsgSend", true},
		{"/cosmos.bank.*", "/cosmos.bank.v1beta1.MsgSend", true},
		{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.*", false},
		{"/cosmos.*", "/cosmos.bank.*", true},
		{"*", "*", true},
		// only an exact entry blocks a wrapper
		{exec, exec, true},
		{"*", exec, false},
		{"/cosmos.*", exec, false},
		{"/cosmos.authz.*", exec, false},
		{"/cosmos.authz.v1beta1.*", exec, false},
		{"*", "/cosmos.group.v1.MsgSubmitProposal", false},
		{"/cosmos.group.*", "/cosmos.group.v1.MsgVote", false},
		{"/cosmos.group.v1.*", "/cosmos.group.v1.MsgExec", false},
		{"/cosmos.group.v1.MsgExec", "/cosmos.group.v1.MsgExec", true},
		// other msgs in the wrappers' packages are blocked as usual
		{"*", "/cosmos.authz.v1beta1.MsgGrant", true},
		{"/cosmos.authz.*", "/cosmos.authz.v1beta1.MsgRevoke", true},
		{"/cosmos.group.*", "/cosmos.group.v1.MsgCreateGroup", true},
		// a prefix trip still blocks a prefix below it
		{"*", "/cosmos.authz.*", true},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, types.Blocks(tc.d, tc.p), "Blocks(%q, %q)", tc.d, tc.p)
	}

	// scope checks keep using Covers, so "*" and "/cosmos.authz.*" scopes
	// still cover the wrapper for authorization
	require.True(t, types.Covers("*", exec))
	require.True(t, types.Covers("/cosmos.authz.*", exec))
}

func TestPrefixPatterns(t *testing.T) {
	require.Equal(t,
		[]string{"/cosmos.*", "/cosmos.bank.*", "/cosmos.bank.v1beta1.*"},
		slices.Collect(types.PrefixPatterns("/cosmos.bank.v1beta1.MsgSend")),
	)
	require.Empty(t, slices.Collect(types.PrefixPatterns("/MsgTopLevel")))
	require.Empty(t, slices.Collect(types.PrefixPatterns("")))

	// stopping early yields nothing more
	var first []string
	for p := range types.PrefixPatterns("/cosmos.bank.v1beta1.MsgSend") {
		first = append(first, p)
		break
	}
	require.Equal(t, []string{"/cosmos.*"}, first)
}

func TestBlockingPatterns(t *testing.T) {
	require.Equal(t,
		[]string{"/cosmos.bank.v1beta1.MsgSend", "*", "/cosmos.*", "/cosmos.bank.*", "/cosmos.bank.v1beta1.*"},
		slices.Collect(types.BlockingPatterns("/cosmos.bank.v1beta1.MsgSend")),
	)
	require.Equal(t, []string{"/MsgTopLevel", "*"}, slices.Collect(types.BlockingPatterns("/MsgTopLevel")))
	for _, w := range []string{
		"/cosmos.authz.v1beta1.MsgExec",
		"/cosmos.group.v1.MsgSubmitProposal",
		"/cosmos.group.v1.MsgVote",
		"/cosmos.group.v1.MsgExec",
	} {
		require.Equal(t, []string{w}, slices.Collect(types.BlockingPatterns(w)), w)
	}

	// stopping early yields nothing more
	var first []string
	for p := range types.BlockingPatterns("/cosmos.bank.v1beta1.MsgSend") {
		first = append(first, p)
		if len(first) == 2 {
			break
		}
	}
	require.Equal(t, []string{"/cosmos.bank.v1beta1.MsgSend", "*"}, first)
}

func TestValidatePatternLength(t *testing.T) {
	longest := "/" + strings.Repeat("a", types.MaxPatternLength-1)
	require.NoError(t, types.ValidatePattern(longest))

	err := types.ValidatePattern(longest + "a")
	require.ErrorIs(t, err, types.ErrInvalidPattern)
	require.Less(t, len(err.Error()), 100, "the error must not echo an oversized input")
}

func TestValidateTypeURL(t *testing.T) {
	require.NoError(t, types.ValidateTypeURL("/cosmos.bank.v1beta1.MsgSend"))
	require.NoError(t, types.ValidateTypeURL("/MsgTopLevel"))
	for _, u := range []string{
		"", "*", "/", "/cosmos.bank.*", "cosmos.bank.v1beta1.MsgSend", "/cosmos..MsgSend",
		"/" + strings.Repeat("a.", types.MaxPatternLength) + "MsgX",
	} {
		require.ErrorIs(t, types.ValidateTypeURL(u), types.ErrInvalidPattern, "%.40q", u)
	}
}
