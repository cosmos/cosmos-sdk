package safety_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/rand/v2"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/msgguard"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// This file fuzzes the msg server guard against the router: random sequences
// of trips, resets and msgs run on one branch of the block state, and every
// msg runs twice from the same state, once through the MsgServiceRouter and
// once as a direct msg server call, as an EVM precompile makes it. The two
// runs must agree on success, on the circuit breaker verdict, on the error
// code, on the events and on every store write. Each step keeps the writes
// of one of the two runs, picked at random, so routed and direct writes
// interleave.

// directFuzz is a router fixture shared by every fuzz input, with one
// proposal in voting so gov MsgVote and MsgDeposit can succeed. Inputs only
// run on branches of the block state and never commit.
type directFuzz struct {
	f        *routerFixture
	methods  map[string]msgguard.Method
	proposal uint64
	keys     []*storetypes.KVStoreKey
}

var (
	directFuzzMu  sync.Mutex
	directFuzzEnv *directFuzz
)

func getDirectFuzz(t *testing.T) *directFuzz {
	t.Helper()
	directFuzzMu.Lock()
	defer directFuzzMu.Unlock()
	if directFuzzEnv != nil {
		directFuzzEnv.f.t = t
		return directFuzzEnv
	}
	f := newRouterFixture(t, false)
	env := &directFuzz{f: f, methods: map[string]msgguard.Method{}}
	for _, m := range msgguard.Methods(t, codec.NewProtoCodec(f.registry), f.app.ModuleManager) {
		env.methods[m.TypeURL] = m
	}
	env.proposal = f.submitProposal(f.msgDistrParams("0.2"))
	for _, k := range f.app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			env.keys = append(env.keys, kv)
		}
	}
	directFuzzEnv = env
	return env
}

// fuzzTripPatterns are the patterns the operator, whose scope is "*", trips
// and resets. Protected patterns are among them, so some trips must fail.
var fuzzTripPatterns = []string{
	types.Wildcard,
	"/cosmos.*",
	"/cosmos.bank.*",
	"/cosmos.bank.v1beta1.*",
	urlSend,
	urlMultiSend,
	"/cosmos.staking.*",
	urlDelegate,
	"/cosmos.staking.v1beta1.MsgUndelegate",
	"/cosmos.distribution.*",
	urlWithdrawReward,
	urlDistrParams,
	"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
	"/cosmos.distribution.v1beta1.MsgFundCommunityPool",
	"/cosmos.authz.*",
	urlAuthzExec,
	urlAuthzGrant,
	urlAuthzRevoke,
	"/cosmos.gov.v1.MsgVote",
}

// msgs returns the msg pool: plain msgs, protected msgs, and authz MsgExec
// wrapping each kind, executed by the grantee under a grant that may or may
// not exist, or by the user on its own behalf.
func (e *directFuzz) msgs() []sdk.Msg {
	f := e.f
	val := f.validator().String()
	one := sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)
	grant, err := authz.NewMsgGrant(f.user.addr, f.grantee.addr, authz.NewGenericAuthorization(urlSend), nil)
	require.NoError(f.t, err)
	revoke := authz.NewMsgRevoke(f.user.addr, f.grantee.addr, urlSend)
	vote := govv1.NewMsgVote(f.user.addr, e.proposal, govv1.OptionYes, "")
	trip := types.NewMsgTrip(f.user.String(), urlMultiSend)
	plain := []sdk.Msg{
		f.msgSend(),
		f.msgMultiSend(),
		f.msgDelegate(),
		stakingtypes.NewMsgUndelegate(f.user.String(), val, one),
		distrtypes.NewMsgWithdrawDelegatorReward(f.user.String(), val),
		distrtypes.NewMsgSetWithdrawAddress(f.user.addr, f.grantee.addr),
		distrtypes.NewMsgFundCommunityPool(oneStake, f.user.String()),
		f.msgDistrParams("0.1"),
		grant,
		&revoke,
		vote,
		// gov rejects deposits below its minimum deposit ratio
		govv1.NewMsgDeposit(f.user.addr, e.proposal, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000))),
	}
	execByGrantee := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend()})
	execSelf := authz.NewMsgExec(f.user.addr, []sdk.Msg{f.msgSend(), f.msgDelegate()})
	execVote := authz.NewMsgExec(f.user.addr, []sdk.Msg{vote})
	execTrip := authz.NewMsgExec(f.user.addr, []sdk.Msg{trip})
	return append(plain, &execByGrantee, &execSelf, &execVote, &execTrip)
}

// outcome is what one run of a msg produced.
type outcome struct {
	err     error
	blocked bool
	events  sdk.Events
	digest  [32]byte
}

// digest hashes every KV pair of every module store visible from ctx.
func (e *directFuzz) digest(ctx sdk.Context) [32]byte {
	h := sha256.New()
	for _, k := range e.keys {
		h.Write([]byte(k.Name()))
		it := ctx.MultiStore().GetKVStore(k).Iterator(nil, nil)
		for ; it.Valid(); it.Next() {
			h.Write(it.Key())
			h.Write([]byte{0})
			h.Write(it.Value())
			h.Write([]byte{1})
		}
		it.Close()
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// routed runs msg through the router handler, as a tx or a gov proposal
// does. The marker on the incoming Context, if any, must not matter.
func (e *directFuzz) routed(ctx sdk.Context, msg sdk.Msg) outcome {
	handler := e.f.app.MsgServiceRouter().Handler(msg)
	require.NotNil(e.f.t, handler, sdk.MsgTypeURL(msg))
	res, err := handler(ctx, msg)
	o := outcome{err: err, blocked: errors.Is(err, types.ErrMsgDisabled)}
	if err == nil {
		o.events = res.GetEvents()
		o.digest = e.digest(ctx)
	}
	return o
}

// direct calls msg's msg server directly. goCtx wraps the sdk.Context in a
// plain context.Context when set, the way some callers pass it.
func (e *directFuzz) direct(ctx sdk.Context, msg sdk.Msg, goCtx bool) outcome {
	m, ok := e.methods[sdk.MsgTypeURL(msg)]
	require.True(e.f.t, ok, "no msg server for %s", sdk.MsgTypeURL(msg))
	var c context.Context = ctx
	if goCtx {
		c = context.WithValue(context.Background(), sdk.SdkContextKey, ctx)
	}
	_, err := m.Call(c, msg)
	o := outcome{err: err, blocked: errors.Is(err, types.ErrMsgDisabled)}
	if err == nil {
		o.events = ctx.EventManager().Events()
		o.digest = e.digest(ctx)
	}
	return o
}

// verdict reports whether the x/safety keeper blocks msg itself and, for
// authz MsgExec, whether it blocks any msg inside it.
func (e *directFuzz) verdict(ctx sdk.Context, msg sdk.Msg) (outer, inner bool) {
	outer = !keeperAllows(e.f.t, e.f, ctx, sdk.MsgTypeURL(msg))
	if exec, ok := msg.(*authz.MsgExec); ok {
		msgs, err := exec.GetMessages()
		require.NoError(e.f.t, err)
		for _, m := range msgs {
			inner = inner || !keeperAllows(e.f.t, e.f, ctx, sdk.MsgTypeURL(m))
		}
	}
	return outer, inner
}

func keeperAllows(t *testing.T, f *routerFixture, ctx sdk.Context, url string) bool {
	t.Helper()
	ok, err := f.safetyKeeper.IsAllowed(ctx, url)
	if errors.Is(err, types.ErrMsgDisabled) {
		require.False(t, ok)
		return false
	}
	require.NoError(t, err)
	return ok
}

// fuzzBytes hands out fuzz bytes, returning 0 once the input runs out.
type fuzzBytes struct{ data []byte }

func (r *fuzzBytes) next() byte {
	if len(r.data) == 0 {
		return 0
	}
	b := r.data[0]
	r.data = r.data[1:]
	return b
}

func (r *fuzzBytes) pick(n int) int { return int(r.next()) % n }

// fuzzStats counts what the steps of a run covered.
type fuzzStats map[string]int

const maxDirectFuzzSteps = 32

// runDirectFuzz runs one random sequence decoded from data and returns what
// it covered.
func runDirectFuzz(t *testing.T, data []byte) fuzzStats {
	t.Helper()
	e := getDirectFuzz(t)
	f := e.f
	pool := e.msgs()
	poolURLs := make([]string, 0, len(pool))
	for _, m := range pool {
		poolURLs = append(poolURLs, sdk.MsgTypeURL(m))
	}

	r := &fuzzBytes{data: data}
	stats := fuzzStats{}
	state := f.blockCtx()

	for step := 0; step < maxDirectFuzzSteps && len(r.data) > 0; step++ {
		var msg sdk.Msg
		switch op := r.pick(8); op {
		case 0, 1:
			msg = types.NewMsgTrip(f.user.String(), fuzzTripPatterns[r.pick(len(fuzzTripPatterns))])
		case 2:
			p := fuzzTripPatterns[r.pick(len(fuzzTripPatterns))]
			if disabled := disabledPatterns(t, f, state); len(disabled) > 0 && r.next()&1 == 1 {
				p = disabled[r.pick(len(disabled))]
			}
			msg = types.NewMsgReset(f.user.String(), p)
		default:
			msg = pool[r.pick(len(pool))]
		}
		url := sdk.MsgTypeURL(msg)
		label := url

		// A stale mark for another msg, as an outer routed msg leaves for
		// a precompile it calls, must not let either run skip its check.
		mode := r.next()
		stale := ""
		if mode&2 == 2 {
			stale = poolURLs[int(mode>>3)%len(poolURLs)]
			if stale == url {
				stale = ""
			}
		}
		outerBlocked, innerBlocked := e.verdict(state, msg)

		routedCtx, writeRouted := state.CacheContext()
		directCtx, writeDirect := state.CacheContext()
		routed := e.routed(routedCtx.WithCheckedMsgTypeURL(stale), msg)
		direct := e.direct(directCtx.WithCheckedMsgTypeURL(stale), msg, mode&4 == 4)

		// A blocked inner msg fails the exec when authz dispatches it, unless
		// a missing grant fails the exec first.
		switch {
		case outerBlocked:
			require.True(t, routed.blocked, "step %d %s: routed run not blocked, err=%v", step, label, routed.err)
		case !innerBlocked:
			require.False(t, routed.blocked, "step %d %s: routed run blocked, err=%v", step, label, routed.err)
		}
		require.Equal(t, routed.err == nil, direct.err == nil, "step %d %s: routed err=%v direct err=%v", step, label, routed.err, direct.err)
		require.Equal(t, routed.blocked, direct.blocked, "step %d %s: routed err=%v direct err=%v", step, label, routed.err, direct.err)
		if routed.err != nil {
			rs, rc, _ := errorsmod.ABCIInfo(routed.err, false)
			ds, dc, _ := errorsmod.ABCIInfo(direct.err, false)
			require.Equal(t, rs+"/"+strconv.FormatUint(uint64(rc), 10), ds+"/"+strconv.FormatUint(uint64(dc), 10), "step %d %s: routed err=%v direct err=%v", step, label, routed.err, direct.err)
		} else {
			require.Equal(t, routed.events, direct.events, "step %d %s: events differ", step, label)
			require.Equal(t, routed.digest, direct.digest, "step %d %s: store writes differ", step, label)
		}

		switch {
		case routed.blocked:
			stats["blocked"]++
		case routed.err == nil:
			stats["ok"]++
		default:
			stats["failed"]++
		}
		if !outerBlocked && innerBlocked && routed.blocked {
			stats["blocked inside exec"]++
		}
		if routed.blocked && types.IsProtected(url) {
			stats["protected blocked"]++
		}

		// Keep one run's writes, so later steps start from a state that
		// routed and direct calls built together.
		if routed.err == nil {
			if mode&1 == 1 {
				writeRouted()
				stats["kept routed"]++
			} else {
				writeDirect()
				stats["kept direct"]++
			}
		}
	}
	return stats
}

func disabledPatterns(t *testing.T, f *routerFixture, ctx sdk.Context) []string {
	t.Helper()
	var out []string
	require.NoError(t, f.safetyKeeper.DisabledList.Walk(ctx, nil, func(p string) (bool, error) {
		out = append(out, p)
		return false, nil
	}))
	return out
}

// FuzzDirectAndRoutedAgree checks that a direct msg server call and a router
// dispatch of the same msg from the same state always agree, under random
// trips and resets.
func FuzzDirectAndRoutedAgree(f *testing.F) {
	f.Add([]byte{})
	// Trip "*" and send through every path.
	f.Add([]byte{0, 0, 1, 3, 0, 1, 3, 1, 4, 3, 12, 7, 3, 13, 5, 3, 15, 0})
	// Trip MsgSend, then exec it inside authz, then reset and retry.
	f.Add([]byte{0, 4, 1, 3, 12, 0, 3, 14, 4, 2, 4, 1, 3, 12, 1})
	// Trip authz MsgExec by its exact URL, vote through it, reset.
	f.Add([]byte{1, 15, 0, 3, 14, 2, 3, 10, 1, 2, 15, 1, 3, 14, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		runDirectFuzz(t, data)
	})
}

// TestDirectAndRoutedAgreeRandomWalk runs the fuzz body over deterministic
// random inputs and checks that the runs reach blocked, failed and
// successful msgs, and msgs blocked inside authz MsgExec, so the fuzz target
// is meaningful.
func TestDirectAndRoutedAgreeRandomWalk(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	total := fuzzStats{}
	for range 120 {
		data := make([]byte, 3*maxDirectFuzzSteps)
		for i := range data {
			data[i] = byte(rng.Uint32())
		}
		for k, v := range runDirectFuzz(t, data) {
			total[k] += v
		}
	}
	t.Logf("outcomes: %v", total)
	for _, k := range []string{"blocked", "ok", "failed", "blocked inside exec", "kept routed", "kept direct"} {
		require.Greater(t, total[k], 10, "%s is too rare for the fuzz to be meaningful", k)
	}
	require.Zero(t, total["protected blocked"], "a protected msg was blocked")
}
