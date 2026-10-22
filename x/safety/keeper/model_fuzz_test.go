package keeper_test

import (
	"context"
	"errors"
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/safety/keeper"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// This file runs a model-based fuzz of the keeper: random sequences of
// UpdateParams, Trip and Reset from random senders run against both the
// keeper and a small reference model, with the module invariants asserted
// after every step. The default test run replays only the seed corpus. To
// fuzz, run go test -run '^$' -fuzz FuzzKeeperModel; the random walk in
// model_walk_test.go runs with -tags sims.

type modelEnv struct {
	cdc      codec.Codec
	registry codectypes.InterfaceRegistry
	ac       address.Codec
	msgURLs  []string
	gov      string
	addrs    []string
}

var getModelEnv = sync.OnceValue(func() modelEnv {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	govv1.RegisterInterfaces(encCfg.InterfaceRegistry)
	authz.RegisterInterfaces(encCfg.InterfaceRegistry)
	consensustypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	upgradetypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	// enterprise/group is its own module, so stand-ins register the group
	// wrapper type URLs. No other group msg is registered, so a group prefix
	// blocks no registered msg and a trip of one must fail.
	custom, ok := encCfg.InterfaceRegistry.(interface {
		RegisterCustomTypeURL(iface any, typeURL string, impl proto.Message)
	})
	if !ok {
		panic("interface registry cannot register custom type URLs")
	}
	for _, url := range []string{"/cosmos.group.v1.MsgSubmitProposal", "/cosmos.group.v1.MsgVote", "/cosmos.group.v1.MsgExec"} {
		custom.RegisterCustomTypeURL((*sdk.Msg)(nil), url, &testdata.TestMsg{})
	}

	gov := authtypes.NewModuleAddress("gov").String()
	op1 := sdk.AccAddress("operator____________").String()
	return modelEnv{
		cdc:      encCfg.Codec,
		registry: encCfg.InterfaceRegistry,
		ac:       addresscodec.NewBech32Codec("cosmos"),
		msgURLs:  encCfg.InterfaceRegistry.ListImplementations(sdk.MsgInterfaceProtoName),
		gov:      gov,
		addrs: []string{
			gov,
			sdk.AccAddress("alt_authority_______").String(),
			op1,
			sdk.AccAddress("operator2___________").String(),
			sdk.AccAddress("stranger____________").String(),
			strings.ToUpper(gov),
			strings.ToUpper(op1),
			"",
			"cosmos1bad",
		},
	}
})

var modelPatterns = []string{
	"/cosmos.bank.v1beta1.MsgSend",
	"/cosmos.bank.v1beta1.MsgMultiSend",
	"/cosmos.bank.v1beta1.*",
	"/cosmos.bank.*",
	"/cosmos.*",
	"*",
	"/cosmos.staking.v1beta1.MsgDelegate",
	"/cosmos.staking.*",
	"/cosmos.gov.v1.MsgVote",
	"/cosmos.gov.*",
	"/cosmos.safety.v1.MsgReset",
	"/cosmos.safety.*",
	"/cosmos.consensus.v1.MsgUpdateParams",
	"/cosmos.consensus.*",
	"/cosmos.upgrade.*",
	"/cosmos.authz.v1beta1.MsgExec",
	"/cosmos.bank.v1beta1.MsgSendX",
	"/cosmos.banking.*",
	"/cosmos.ban.*",
	"/ibc.*",
	"bad",
	"/cosmos.bank.*.*",
	"",
	"/cosmos.authz.*",
	"/cosmos.authz.v1beta1.MsgGrant",
	"/cosmos.group.v1.MsgVote",
	"/cosmos.group.*",
	"/cosmos.group.v1.*",
	"/cosmos.group.v1.MsgExec",
	"/cosmos.group.v1.MsgSubmitProposal",
}

var modelScopes = [][]string{
	nil,
	{"*"},
	{"/cosmos.bank.*"},
	{"/cosmos.bank.v1beta1.MsgSend"},
	{"/cosmos.staking.*", "/cosmos.bank.v1beta1.*"},
	{"/cosmos.*"},
	{"/cosmos.gov.*"},
	{"/cosmos.banking.*"},
	{"*", "*"},
	{"bad"},
	{"/cosmos.authz.*"},
	{"/cosmos.authz.v1beta1.MsgExec"},
	{"/cosmos.group.*"},
}

// Reference model. It shares no code with the keeper or the types package.

var modelPatternRE = regexp.MustCompile(`^(\*|/[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*(\.\*)?)$`)

func modelCovers(s, p string) bool {
	if s == "*" || s == p {
		return true
	}
	if p == "*" || !strings.HasSuffix(s, ".*") {
		return false
	}
	sSegs := strings.Split(strings.TrimSuffix(s[1:], ".*"), ".")
	pPrefix := strings.HasSuffix(p, ".*")
	pSegs := strings.Split(strings.TrimSuffix(p[1:], ".*"), ".")
	if len(sSegs) > len(pSegs) || (!pPrefix && len(sSegs) == len(pSegs)) {
		return false
	}
	return slices.Equal(sSegs, pSegs[:len(sSegs)])
}

func modelProtected(url string) bool {
	return url == "/cosmos.consensus.v1.MsgUpdateParams" ||
		strings.HasPrefix(url, "/cosmos.safety.") ||
		strings.HasPrefix(url, "/cosmos.gov.") ||
		strings.HasPrefix(url, "/cosmos.upgrade.")
}

func modelWrapper(url string) bool {
	switch url {
	case "/cosmos.authz.v1beta1.MsgExec", "/cosmos.group.v1.MsgSubmitProposal",
		"/cosmos.group.v1.MsgVote", "/cosmos.group.v1.MsgExec":
		return true
	}
	return false
}

// modelBlocks reports whether disabled entry d blocks pattern p. Only an
// exact entry blocks a wrapper.
func modelBlocks(d, p string) bool {
	if modelWrapper(p) {
		return d == p
	}
	return modelCovers(d, p)
}

type model struct {
	env      modelEnv
	params   types.Params
	disabled map[string]bool
	// override is the consensus AuthorityParams authority, or "" for none.
	override string
}

func (m *model) validAddr(a string) bool {
	if a == "" {
		return false
	}
	_, err := m.env.ac.StringToBytes(a)
	return err == nil
}

func (m *model) validPatterns(ps []string, allowEmpty bool) bool {
	if len(ps) == 0 {
		return allowEmpty
	}
	seen := map[string]bool{}
	for _, p := range ps {
		if !modelPatternRE.MatchString(p) || seen[p] {
			return false
		}
		seen[p] = true
	}
	return true
}

// canonicalAddr reports whether a decodes and re-encodes to itself. Params
// addresses and the trip operator must be canonical; the reset and
// UpdateParams senders need only decode, and match roles by exact string.
func (m *model) canonicalAddr(a string) bool {
	bz, err := m.env.ac.StringToBytes(a)
	if err != nil {
		return false
	}
	s, err := m.env.ac.BytesToString(bz)
	return err == nil && s == a
}

func (m *model) validParams(p types.Params) bool {
	return m.validAddr(p.Authority) && m.canonicalAddr(p.Authority) &&
		(p.Operator == "" || m.canonicalAddr(p.Operator)) &&
		m.validPatterns(p.Scope, true)
}

func (m *model) authority() string {
	if m.override != "" {
		return m.override
	}
	return m.params.Authority
}

// isOperator matches by exact string: params.operator is canonical, and a
// trip sender must be too.
func (m *model) isOperator(sender string) bool {
	return m.params.Operator != "" && sender == m.params.Operator
}

func (m *model) inScope(ps []string) bool {
	for _, p := range ps {
		covered := false
		for _, s := range m.params.Scope {
			covered = covered || modelCovers(s, p)
		}
		if !covered {
			return false
		}
	}
	return true
}

func (m *model) matchesUnprotected(p string) bool {
	for _, url := range m.env.msgURLs {
		if !modelProtected(url) && modelBlocks(p, url) {
			return true
		}
	}
	return false
}

func (m *model) allowed(url string) bool {
	if modelProtected(url) {
		return true
	}
	for d := range m.disabled {
		if modelBlocks(d, url) {
			return false
		}
	}
	return true
}

func (m *model) updateParams(sender string, p types.Params) bool {
	if !m.validAddr(sender) || !m.validParams(p) || sender != m.authority() {
		return false
	}
	m.params = p
	return true
}

func (m *model) trip(sender string, ps []string) bool {
	if !m.validAddr(sender) || !m.validPatterns(ps, false) || !m.isOperator(sender) || !m.inScope(ps) {
		return false
	}
	for _, p := range ps {
		if !m.matchesUnprotected(p) || m.disabled[p] {
			return false
		}
	}
	for _, p := range ps {
		m.disabled[p] = true
	}
	return true
}

func (m *model) reset(sender string, ps []string) bool {
	if !m.validAddr(sender) || !m.validPatterns(ps, false) {
		return false
	}
	if sender != m.authority() && (!m.isOperator(sender) || !m.inScope(ps)) {
		return false
	}
	for _, p := range ps {
		if !m.disabled[p] {
			return false
		}
	}
	for _, p := range ps {
		delete(m.disabled, p)
	}
	return true
}

// System under test.

type modelSUT struct {
	t   *testing.T
	ctx sdk.Context
	k   keeper.Keeper
	srv types.MsgServer
}

func newModelSUT(t *testing.T, env modelEnv) modelSUT {
	t.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})
	k := keeper.NewKeeper(env.cdc, runtime.NewKVStoreService(key), env.ac, env.registry, env.gov)
	return modelSUT{t: t, ctx: ctx, k: k, srv: keeper.NewMsgServerImpl(k)}
}

func (s modelSUT) withOverride(override string) sdk.Context {
	cp := s.ctx.ConsensusParams()
	if override == "" {
		cp.Authority = nil
	} else {
		cp.Authority = &cmtproto.AuthorityParams{Authority: override}
	}
	return s.ctx.WithConsensusParams(cp).WithEventManager(sdk.NewEventManager())
}

func (s modelSUT) disabled(ctx sdk.Context) []string {
	var out []string
	require.NoError(s.t, s.k.DisabledList.Walk(ctx, nil, func(p string) (bool, error) {
		out = append(out, p)
		return false, nil
	}))
	return out
}

// byteReader hands out fuzz bytes, returning 0 once the input runs out.
type byteReader struct{ data []byte }

func (r *byteReader) next() byte {
	if len(r.data) == 0 {
		return 0
	}
	b := r.data[0]
	r.data = r.data[1:]
	return b
}

func (r *byteReader) pick(n int) int { return int(r.next()) % n }

func (r *byteReader) patterns(m *model) []string {
	n := r.pick(4)
	ps := make([]string, 0, n)
	current := slices.Sorted(maps.Keys(m.disabled))
	for range n {
		if b := r.next(); b&1 == 1 && len(current) > 0 {
			ps = append(ps, current[int(b>>1)%len(current)])
		} else {
			ps = append(ps, modelPatterns[int(b>>1)%len(modelPatterns)])
		}
	}
	return ps
}

func (r *byteReader) params(env modelEnv) types.Params {
	op := ""
	if b := r.next(); b%4 != 0 {
		op = env.addrs[int(b>>2)%len(env.addrs)]
	}
	return types.Params{
		Authority: env.addrs[r.pick(len(env.addrs))],
		Operator:  op,
		Scope:     slices.Clone(modelScopes[r.pick(len(modelScopes))]),
	}
}

const maxModelSteps = 48

// FuzzKeeperModel runs a random genesis and a random sequence of msgs against
// the keeper and the reference model.
func FuzzKeeperModel(f *testing.F) {
	f.Add([]byte{})
	// Operator trips bank, authority resets it.
	f.Add([]byte{1, 2, 0, 0, 2, 2, 1, 0, 5, 0, 1, 1})
	// Genesis disables the reset msg and "*"; authority resets them.
	f.Add([]byte{3, 0x0b, 0x15, 2, 4, 5, 0, 2, 1, 1, 5, 0, 2, 1, 3})
	// Params churn, override toggles, out-of-scope trips.
	f.Add([]byte{0, 0, 9, 1, 3, 7, 1, 2, 3, 2, 6, 6, 7, 3, 2, 8, 1, 4, 5, 2, 2, 7, 0, 5, 5, 2, 1, 1})
	// Under scope "*", the operator trips "*" and authz MsgExec in one msg,
	// then resets "*" and MsgExec, in each order.
	f.Add([]byte{1, 0, 2, 1, 2, 10, 30, 5, 1, 1, 1, 5, 1, 1, 1})
	f.Add([]byte{1, 0, 2, 1, 2, 30, 10, 5, 1, 1, 3, 5, 1, 1, 1})
	// Under scope "*", group prefix trips fail since they block no
	// registered msg, group MsgExec trips by its exact URL, and "*" trips
	// beside group MsgSubmitProposal.
	f.Add([]byte{1, 0, 2, 1, 1, 52, 2, 1, 1, 54, 2, 1, 1, 56, 2, 1, 2, 10, 58})

	f.Fuzz(func(t *testing.T, data []byte) {
		runKeeperModel(t, getModelEnv(), data)
	})
}

// runKeeperModel runs one random genesis and msg sequence decoded from data
// and returns how many msgs of each kind succeeded.
func runKeeperModel(t *testing.T, env modelEnv, data []byte) map[string]int {
	t.Helper()
	r := &byteReader{data: data}
	sut := newModelSUT(t, env)
	m := &model{env: env, disabled: map[string]bool{}}
	stats := map[string]int{}

	// Random genesis: any well-formed patterns, including protected
	// ones, may start disabled.
	gp := types.Params{Authority: env.gov, Operator: env.addrs[2], Scope: slices.Clone(modelScopes[r.pick(6)])}
	var gd []string
	for range r.pick(4) {
		p := modelPatterns[r.pick(len(modelPatterns))]
		if modelPatternRE.MatchString(p) && !slices.Contains(gd, p) {
			gd = append(gd, p)
		}
	}
	require.NoError(t, sut.k.InitGenesis(sut.ctx, types.NewGenesisState(gp, gd)))
	m.params = gp
	for _, p := range gd {
		m.disabled[p] = true
	}
	checkModelInvariants(t, sut, m, "genesis")

	for step := 0; step < maxModelSteps && len(r.data) > 0; step++ {
		op := r.pick(8)
		// Half the time the sender is the role that can act, so that msgs
		// succeed often enough to reach interesting states.
		sb := r.next()
		sender := env.addrs[int(sb>>1)%len(env.addrs)]
		if sb&1 == 1 {
			if op <= 1 || (op >= 5 && sb&2 == 2) {
				sender = m.authority()
			} else if m.params.Operator != "" {
				sender = m.params.Operator
			}
		}
		ctx := sut.withOverride(m.override)
		before := *m
		before.params.Scope = slices.Clone(m.params.Scope)

		var (
			desc      string
			wantOK    bool
			err       error
			eventType string
			patterns  []string
		)
		switch op {
		case 0, 1:
			p := r.params(env)
			desc = "UpdateParams"
			wantOK = m.updateParams(sender, p)
			_, err = sut.srv.UpdateParams(ctx, &types.MsgUpdateParams{Authority: sender, Params: p})
			eventType = "cosmos.safety.v1.EventParamsUpdated"
		case 2, 3, 4:
			patterns = r.patterns(m)
			desc = "Trip"
			wantOK = m.trip(sender, patterns)
			_, err = sut.srv.Trip(ctx, types.NewMsgTrip(sender, patterns...))
			eventType = "cosmos.safety.v1.EventTrip"
		case 5, 6:
			patterns = r.patterns(m)
			desc = "Reset"
			wantOK = m.reset(sender, patterns)
			_, err = sut.srv.Reset(ctx, types.NewMsgReset(sender, patterns...))
			eventType = "cosmos.safety.v1.EventReset"
		case 7:
			// Toggle the consensus AuthorityParams override.
			if b := r.next(); b%3 == 0 {
				m.override = ""
			} else {
				m.override = env.addrs[int(b)%7]
			}
			checkModelInvariants(t, sut, m, "override")
			continue
		}

		if err == nil {
			stats[desc]++
		}
		label := desc + " sender=" + sender + " patterns=" + strings.Join(patterns, ",")
		if wantOK {
			require.NoError(t, err, "step %d %s: model accepted", step, label)
			evs := ctx.EventManager().Events()
			require.NotEmpty(t, evs, "step %d %s: no event", step, label)
			require.Equal(t, eventType, evs[len(evs)-1].Type, "step %d %s", step, label)
		} else {
			require.Error(t, err, "step %d %s: model rejected", step, label)
			require.Empty(t, ctx.EventManager().Events(), "step %d %s: event on failure", step, label)
		}

		// The operator never acts outside its scope: a successful trip or
		// reset by anyone but the authority uses only in-scope patterns.
		if err == nil && (desc == "Trip" || (desc == "Reset" && sender != before.authority())) {
			require.True(t, before.isOperator(sender), "step %d %s: non-operator acted", step, label)
			require.True(t, before.inScope(patterns), "step %d %s: out of scope", step, label)
		}
		checkModelInvariants(t, sut, m, label)
	}
	return stats
}

// checkModelInvariants asserts that the keeper matches the model and checks
// the module invariants.
func checkModelInvariants(t *testing.T, sut modelSUT, m *model, label string) {
	t.Helper()
	ctx := sut.withOverride(m.override)

	// State matches the model.
	params, err := sut.k.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, m.params.Authority, params.Authority, label)
	require.Equal(t, m.params.Operator, params.Operator, label)
	require.Equal(t, nilIfEmpty(m.params.Scope), nilIfEmpty(params.Scope), label)
	require.Equal(t, slices.Sorted(maps.Keys(m.disabled)), nilIfEmpty(sut.disabled(ctx)), label)

	// Protected URLs are always allowed, and IsAllowed agrees with the
	// model for every registered msg.
	urls := append(slices.Clone(m.env.msgURLs),
		"/cosmos.safety.v1.MsgReset", "/cosmos.gov.v1.MsgVote", "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		"/cosmos.consensus.v1.MsgUpdateParams", "/cosmos.banking.v1.MsgX", "/ibc.core.v1.MsgY",
		"/cosmos.group.v1.MsgSubmitProposal", "/cosmos.group.v1.MsgVote", "/cosmos.group.v1.MsgExec",
		"/cosmos.group.v1.MsgCreateGroup")
	for _, url := range urls {
		ok := keeperAllowed(t, sut.k, ctx, url)
		if modelProtected(url) {
			require.True(t, ok, "%s: protected %s blocked", label, url)
		}
		if modelWrapper(url) {
			require.Equal(t, !m.disabled[url], ok, "%s: wrapper %s must be blocked by its exact entry only", label, url)
		}
		require.Equal(t, m.allowed(url), ok, "%s: IsAllowed(%s) disabled=%v", label, url, sut.disabled(ctx))
	}

	// The authority can never be locked out of reset: from any state, the
	// effective authority resets each disabled pattern, and all of them at
	// once, in a throwaway branch.
	if len(m.disabled) > 0 {
		auth := m.authority()
		all := slices.Sorted(maps.Keys(m.disabled))
		for _, p := range all {
			cctx, _ := ctx.CacheContext()
			_, err := sut.srv.Reset(cctx, types.NewMsgReset(auth, p))
			require.NoError(t, err, "%s: authority %s cannot reset %s", label, auth, p)
		}
		cctx, _ := ctx.CacheContext()
		_, err := sut.srv.Reset(cctx, types.NewMsgReset(auth, all...))
		require.NoError(t, err, "%s: authority cannot reset all of %v", label, all)
		require.True(t, keeperAllowed(t, sut.k, cctx, "/cosmos.bank.v1beta1.MsgSend"), "%s: MsgSend still blocked after full reset", label)
	}

	// State always passes genesis validation, and export/import round-trips.
	gs, err := sut.k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, gs.Validate(m.env.ac), "%s: exported genesis invalid", label)

	fresh := newModelSUT(t, m.env)
	require.NoError(t, fresh.k.InitGenesis(fresh.ctx, gs), "%s: import", label)
	gs2, err := fresh.k.ExportGenesis(fresh.ctx)
	require.NoError(t, err)
	require.Equal(t, gs.Params.Authority, gs2.Params.Authority, label)
	require.Equal(t, gs.Params.Operator, gs2.Params.Operator, label)
	require.Equal(t, nilIfEmpty(gs.Params.Scope), nilIfEmpty(gs2.Params.Scope), label)
	require.Equal(t, nilIfEmpty(gs.DisabledPatterns), nilIfEmpty(gs2.DisabledPatterns), label)
	freshCtx := fresh.withOverride(m.override)
	for _, url := range urls {
		a := keeperAllowed(t, sut.k, ctx, url)
		b := keeperAllowed(t, fresh.k, freshCtx, url)
		require.Equal(t, a, b, "%s: IsAllowed(%s) differs after import", label, url)
	}
}

// keeperAllowed calls IsAllowed, which reports a block as false with
// ErrMsgDisabled, and fails on any other error.
func keeperAllowed(t require.TestingT, k keeper.Keeper, ctx context.Context, url string) bool {
	ok, err := k.IsAllowed(ctx, url)
	if errors.Is(err, types.ErrMsgDisabled) {
		require.False(t, ok)
		return false
	}
	require.NoError(t, err)
	return ok
}

func nilIfEmpty(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
