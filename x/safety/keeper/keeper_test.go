package keeper_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	corestore "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/safety/keeper"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// Type URLs used across the unit suite. SetupTest registers every one of
// them. The IsAllowed table also uses unregistered look-alikes, such as
// "/cosmos.banking.v1.MsgX", written inline.
const (
	uSend         = "/cosmos.bank.v1beta1.MsgSend"
	uMultiSend    = "/cosmos.bank.v1beta1.MsgMultiSend"
	uDelegate     = "/cosmos.staking.v1beta1.MsgDelegate"
	uUndelegate   = "/cosmos.staking.v1beta1.MsgUndelegate"
	uExec         = "/cosmos.authz.v1beta1.MsgExec"
	uVote         = "/cosmos.gov.v1.MsgVote"
	uVoteBeta     = "/cosmos.gov.v1beta1.MsgVote"
	uSubmitProp   = "/cosmos.gov.v1.MsgSubmitProposal"
	uSafetyReset  = "/cosmos.safety.v1.MsgReset"
	uSafetyTrip   = "/cosmos.safety.v1.MsgTrip"
	uSafetyParams = "/cosmos.safety.v1.MsgUpdateParams"
	uUpgrade      = "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade"
	uConsensus    = "/cosmos.consensus.v1.MsgUpdateParams"
)

// KeeperTestSuite is a table-driven suite for the x/safety keeper, msg server
// and query server.
type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	key         *storetypes.KVStoreKey
	registry    codectypes.InterfaceRegistry
	keeper      keeper.Keeper
	msgServer   types.MsgServer
	queryServer types.QueryServer
	queryClient types.QueryClient

	authority string
	operator  string
	stranger  string
	other     string
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.key = storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), s.key, storetypes.NewTransientStoreKey("transient_unit"))
	s.ctx = testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1}).WithGasMeter(storetypes.NewInfiniteGasMeter())

	encCfg := moduletestutil.MakeTestEncodingConfig()
	s.registry = encCfg.InterfaceRegistry
	types.RegisterInterfaces(s.registry)
	banktypes.RegisterInterfaces(s.registry)
	stakingtypes.RegisterInterfaces(s.registry)
	govv1.RegisterInterfaces(s.registry)
	govv1beta1.RegisterInterfaces(s.registry)
	authz.RegisterInterfaces(s.registry)
	consensustypes.RegisterInterfaces(s.registry)
	upgradetypes.RegisterInterfaces(s.registry)

	s.authority = authtypes.NewModuleAddress("gov").String()
	s.operator = sdk.AccAddress("unit_operator_______").String()
	s.stranger = sdk.AccAddress("unit_stranger_______").String()
	s.other = sdk.AccAddress("unit_other__________").String()

	s.keeper = keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(s.key), addresscodec.NewBech32Codec("cosmos"), s.registry, s.authority)
	s.msgServer = keeper.NewMsgServerForTest(s.keeper)
	s.queryServer = keeper.NewQueryServer(s.keeper)

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.registry)
	types.RegisterQueryServer(queryHelper, s.queryServer)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.Require().NoError(s.keeper.InitGenesis(s.ctx, types.NewGenesisState(
		types.Params{Authority: s.authority, Operator: s.operator, Scope: []string{types.Wildcard}}, nil,
	)))
}

// state replaces the params and disabled list directly in the store.
func (s *KeeperTestSuite) state(operator string, scope []string, disabled ...string) {
	s.Require().NoError(s.keeper.Params.Set(s.ctx, types.Params{Authority: s.authority, Operator: operator, Scope: scope}))
	s.Require().NoError(s.keeper.DisabledList.Clear(s.ctx, nil))
	for _, p := range disabled {
		s.Require().NoError(s.keeper.DisabledList.Set(s.ctx, p))
	}
}

func (s *KeeperTestSuite) list() []string {
	out := []string{}
	s.Require().NoError(s.keeper.DisabledList.Walk(s.ctx, nil, func(p string) (bool, error) {
		out = append(out, p)
		return false, nil
	}))
	return out
}

func (s *KeeperTestSuite) params() types.Params {
	p, err := s.keeper.Params.Get(s.ctx)
	s.Require().NoError(err)
	return p
}

func (s *KeeperTestSuite) allowed(typeURL string) bool {
	return keeperAllowed(s.T(), s.keeper, s.ctx, typeURL)
}

// freshEvents returns s.ctx with an empty event manager.
func (s *KeeperTestSuite) freshEvents() sdk.Context {
	return s.ctx.WithEventManager(sdk.NewEventManager())
}

// typedEvents parses every event emitted on ctx that is an x/safety typed event.
func (s *KeeperTestSuite) typedEvents(ctx sdk.Context) []proto.Message {
	var out []proto.Message
	for _, ev := range ctx.EventManager().Events() {
		msg, err := sdk.ParseTypedEvent(abci.Event(ev))
		if err != nil {
			continue
		}
		// JSON decodes an empty repeated field as an empty slice; compare as nil.
		if e, ok := msg.(*types.EventParamsUpdated); ok && len(e.Params.Scope) == 0 {
			e.Params.Scope = nil
		}
		out = append(out, msg)
	}
	return out
}

// requireRegisteredError checks that err has a registered ABCI code, either an
// x/safety error or a shared SDK error such as sdkerrors.ErrUnauthorized.
func (s *KeeperTestSuite) requireRegisteredError(err error) {
	codespace, code, _ := errorsmod.ABCIInfo(err, false)
	s.Require().Contains([]string{types.ModuleName, sdkerrors.RootCodespace}, codespace, "error %q has an unregistered codespace", err)
	s.Require().Greater(code, uint32(1), "error %q has no registered code", err)
}

// Every type URL constant resolves in the suite's registry.
func (s *KeeperTestSuite) TestTypeURLConstantsRegistered() {
	for _, url := range []string{
		uSend, uMultiSend, uDelegate, uUndelegate, uExec, uVote, uVoteBeta, uSubmitProp,
		uSafetyReset, uSafetyTrip, uSafetyParams, uUpgrade, uConsensus,
	} {
		_, err := s.registry.Resolve(url)
		s.Require().NoError(err, url)
	}
}

func (s *KeeperTestSuite) TestIsAllowedTable() {
	cases := []struct {
		name     string
		disabled []string
		url      string
		want     bool
	}{
		{"empty list allows", nil, uSend, true},

		{"exact blocks itself", []string{uSend}, uSend, false},
		{"exact does not block sibling", []string{uSend}, uMultiSend, true},
		{"exact does not block longer name", []string{uSend}, uSend + "Extra", true},
		{"exact does not block shorter name", []string{"/cosmos.bank.v1beta1.MsgSen"}, uSend, true},
		{"package path as exact does not block", []string{"/cosmos.bank.v1beta1"}, uSend, true},

		{"prefix at first segment", []string{"/cosmos.*"}, uSend, false},
		{"prefix at second segment", []string{"/cosmos.bank.*"}, uSend, false},
		{"prefix at third segment", []string{"/cosmos.bank.v1beta1.*"}, uSend, false},
		{"prefix past last segment does not block", []string{uSend + ".*"}, uSend, true},
		{"prefix stops at segment boundary", []string{"/cosmos.bank.*"}, "/cosmos.banking.v1.MsgX", true},
		{"partial segment prefix does not block", []string{"/cosmos.ban.*"}, uSend, true},
		{"prefix does not block sibling package", []string{"/cosmos.bank.*"}, uDelegate, true},
		{"prefix does not block sibling version", []string{"/cosmos.bank.v1beta1.*"}, "/cosmos.bank.v2.MsgSend", true},
		{"first segment does not block other root", []string{"/cosmos.*"}, "/ibc.core.channel.v1.MsgRecvPacket", true},

		{"wildcard blocks bank", []string{types.Wildcard}, uSend, false},
		{"wildcard blocks staking", []string{types.Wildcard}, uDelegate, false},
		{"wildcard blocks non-cosmos root", []string{types.Wildcard}, "/ibc.core.channel.v1.MsgRecvPacket", false},
		{"wildcard blocks authz grant", []string{types.Wildcard}, "/cosmos.authz.v1beta1.MsgGrant", false},

		{"protected safety reset under wildcard", []string{types.Wildcard}, uSafetyReset, true},
		{"protected safety trip under wildcard", []string{types.Wildcard}, uSafetyTrip, true},
		{"protected safety params under wildcard", []string{types.Wildcard}, uSafetyParams, true},
		{"protected gov v1 under wildcard", []string{types.Wildcard}, uVote, true},
		{"protected gov v1beta1 under wildcard", []string{types.Wildcard}, uVoteBeta, true},
		{"protected gov submit under wildcard", []string{types.Wildcard}, uSubmitProp, true},
		{"protected upgrade under wildcard", []string{types.Wildcard}, uUpgrade, true},
		{"protected consensus params under wildcard", []string{types.Wildcard}, uConsensus, true},
		{"protected under every prefix and exact", []string{"/cosmos.*", "/cosmos.safety.*", "/cosmos.safety.v1.*", uSafetyReset}, uSafetyReset, true},
		{"protected gov under exact entry", []string{uVote, "/cosmos.gov.*"}, uVote, true},
		// only an exact entry blocks a wrapper; "*" and prefix entries do not
		{"authz exec under wildcard", []string{types.Wildcard}, uExec, true},
		{"authz exec under root prefix", []string{"/cosmos.*"}, uExec, true},
		{"authz exec under package prefix", []string{"/cosmos.authz.*"}, uExec, true},
		{"authz exec under version prefix", []string{"/cosmos.authz.v1beta1.*"}, uExec, true},
		{"authz exec under every broad entry", []string{types.Wildcard, "/cosmos.*", "/cosmos.authz.*", "/cosmos.authz.v1beta1.*"}, uExec, true},
		{"authz exec under exact", []string{uExec}, uExec, false},
		{"authz exec under exact and wildcard", []string{uExec, types.Wildcard}, uExec, false},
		{"authz grant under package prefix", []string{"/cosmos.authz.*"}, "/cosmos.authz.v1beta1.MsgGrant", false},
		{"authz exec look-alike under wildcard", []string{types.Wildcard}, uExec + "X", false},
		{"group submit under wildcard", []string{types.Wildcard}, "/cosmos.group.v1.MsgSubmitProposal", true},
		{"group vote under prefix", []string{"/cosmos.group.*"}, "/cosmos.group.v1.MsgVote", true},
		{"group exec under version prefix", []string{"/cosmos.group.v1.*"}, "/cosmos.group.v1.MsgExec", true},
		{"group submit under exact", []string{"/cosmos.group.v1.MsgSubmitProposal"}, "/cosmos.group.v1.MsgSubmitProposal", false},
		{"group vote under exact", []string{"/cosmos.group.v1.MsgVote"}, "/cosmos.group.v1.MsgVote", false},
		{"group exec under exact", []string{"/cosmos.group.v1.MsgExec"}, "/cosmos.group.v1.MsgExec", false},
		{"group vote exact does not block group exec", []string{"/cosmos.group.v1.MsgVote"}, "/cosmos.group.v1.MsgExec", true},
		{"other group msg under prefix", []string{"/cosmos.group.*"}, "/cosmos.group.v1.MsgCreateGroup", false},

		{"governance look-alike not protected", []string{types.Wildcard}, "/cosmos.governance.v1.MsgX", false},
		{"safety look-alike not protected", []string{types.Wildcard}, "/cosmos.safetyx.v1.MsgX", false},
		{"upgrade look-alike not protected", []string{types.Wildcard}, "/cosmos.upgrader.v1.MsgX", false},
		{"other consensus msg not protected", []string{types.Wildcard}, "/cosmos.consensus.v1.MsgOther", false},
		{"consensus params look-alike not protected", []string{types.Wildcard}, uConsensus + "X", false},
		{"non-cosmos gov not protected", []string{types.Wildcard}, "/other.gov.v1.MsgVote", false},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.state(s.operator, []string{types.Wildcard}, tc.disabled...)
			s.Require().Equal(tc.want, s.allowed(tc.url))

			res, err := s.queryClient.IsAllowed(s.ctx, &types.QueryIsAllowedRequest{MsgTypeUrl: tc.url})
			s.Require().NoError(err)
			s.Require().Equal(tc.want, res.Allowed, "query must agree with the keeper")
			if tc.want {
				s.Require().Empty(res.BlockedBy)
			} else {
				s.Require().Contains(tc.disabled, res.BlockedBy)
				s.Require().True(types.Blocks(res.BlockedBy, tc.url), res.BlockedBy)
			}
		})
	}
}

// countingStoreService opens stores that count their Has calls.
type countingStoreService struct {
	inner corestore.KVStoreService
	has   *int
}

func (c countingStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return countingStore{c.inner.OpenKVStore(ctx), c.has}
}

type countingStore struct {
	corestore.KVStore
	has *int
}

func (c countingStore) Has(key []byte) (bool, error) {
	*c.has++
	return c.KVStore.Has(key)
}

// IsAllowed is a bounded number of point lookups: two plus one per package
// segment, independent of the size of the disabled list, and none for a
// protected type URL. It charges no gas for them: a finite gas meter on the
// caller's context stays at zero.
//
// Before the check charged no gas, this test counted lookups by the gas they
// charged (1000 each, 5000 for a MsgSend). It now counts them at the store.
func (s *KeeperTestSuite) TestIsAllowedBoundedLookups() {
	var has int
	k := keeper.NewKeeper(moduletestutil.MakeTestEncodingConfig().Codec, countingStoreService{runtime.NewKVStoreService(s.key), &has}, addresscodec.NewBech32Codec("cosmos"), s.registry, s.authority)
	lookups := func(url string) int {
		has = 0
		ctx := s.ctx.WithGasMeter(storetypes.NewGasMeter(1))
		_, err := k.IsAllowed(ctx, url)
		s.Require().NoError(err)
		s.Require().Zero(ctx.GasMeter().GasConsumed(), "the check must charge no gas")
		return has
	}

	s.state(s.operator, []string{types.Wildcard})
	s.Require().Equal(5, lookups(uSend), "exact + * + 3 prefixes")
	s.Require().Equal(4, lookups("/a.b.MsgC"), "exact + * + 2 prefixes")
	s.Require().Zero(lookups(uSafetyReset), "protected URL needs no lookup")
	s.Require().Equal(1, lookups(uExec), "a wrapper needs only its exact lookup")

	var many []string
	for i := range 500 {
		many = append(many, fmt.Sprintf("/zz.m%03d.MsgX", i))
	}
	s.state(s.operator, []string{types.Wildcard}, many...)
	s.Require().Equal(5, lookups(uSend), "lookups must not grow with the disabled list")
}

// IsAllowed of a blocked msg also charges no gas, and leaves the caller's
// gas meter as it was.
func (s *KeeperTestSuite) TestIsAllowedBlockedChargesNoGas() {
	s.state(s.operator, []string{types.Wildcard}, uSend)
	meter := storetypes.NewGasMeter(10_000)
	meter.ConsumeGas(7, "before")
	ok, err := s.keeper.IsAllowed(s.ctx.WithGasMeter(meter), uSend)
	s.Require().ErrorIs(err, types.ErrMsgDisabled)
	s.Require().False(ok)
	s.Require().Equal(storetypes.Gas(7), meter.GasConsumed())
}

func (s *KeeperTestSuite) TestUpdateParamsTable() {
	newAuth := sdk.AccAddress("unit_new_authority__").String()
	valid := types.Params{Authority: newAuth, Operator: s.other, Scope: []string{"/cosmos.bank.*", uDelegate}}

	cases := []struct {
		name   string
		signer string
		params types.Params
		err    error
	}{
		{"authority sets params", s.authority, valid, nil},
		{"empty scope and no operator", s.authority, types.Params{Authority: s.authority}, nil},
		{"operator equal to authority", s.authority, types.Params{Authority: s.authority, Operator: s.authority, Scope: []string{types.Wildcard}}, nil},
		{"wildcard scope", s.authority, types.Params{Authority: s.authority, Operator: s.operator, Scope: []string{types.Wildcard}}, nil},
		{"authority replaces itself", s.authority, types.Params{Authority: newAuth}, nil},

		{"operator cannot set params", s.operator, valid, sdkerrors.ErrUnauthorized},
		{"stranger cannot set params", s.stranger, valid, sdkerrors.ErrUnauthorized},
		{"empty signer", "", valid, sdkerrors.ErrInvalidAddress},
		{"malformed signer", "cosmos1notanaddress", valid, sdkerrors.ErrInvalidAddress},
		// roles match by exact string; a non-canonical signer that matches no
		// role is told its address is not canonical
		{"uppercase signer", strings.ToUpper(s.authority), valid, sdkerrors.ErrInvalidAddress},
		{"empty new authority", s.authority, types.Params{Operator: s.operator}, types.ErrInvalidParams},
		{"malformed new authority", s.authority, types.Params{Authority: "cosmos1bad"}, sdkerrors.ErrInvalidAddress},
		{"wrong prefix new authority", s.authority, types.Params{Authority: "osmo1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"}, sdkerrors.ErrInvalidAddress},
		{"malformed operator", s.authority, types.Params{Authority: s.authority, Operator: "nope"}, sdkerrors.ErrInvalidAddress},
		{"duplicate scope", s.authority, types.Params{Authority: s.authority, Scope: []string{uSend, uSend}}, types.ErrDuplicatePattern},
		{"scope without slash", s.authority, types.Params{Authority: s.authority, Scope: []string{"cosmos.bank.*"}}, types.ErrInvalidPattern},
		{"scope empty string", s.authority, types.Params{Authority: s.authority, Scope: []string{""}}, types.ErrInvalidPattern},
		{"scope bare slash", s.authority, types.Params{Authority: s.authority, Scope: []string{"/"}}, types.ErrInvalidPattern},
		{"scope slash star", s.authority, types.Params{Authority: s.authority, Scope: []string{"/*"}}, types.ErrInvalidPattern},
		{"scope slash dot star", s.authority, types.Params{Authority: s.authority, Scope: []string{"/.*"}}, types.ErrInvalidPattern},
		{"scope empty segment", s.authority, types.Params{Authority: s.authority, Scope: []string{"/cosmos..bank.*"}}, types.ErrInvalidPattern},
		{"scope trailing dot", s.authority, types.Params{Authority: s.authority, Scope: []string{"/cosmos.bank."}}, types.ErrInvalidPattern},
		{"scope inner star", s.authority, types.Params{Authority: s.authority, Scope: []string{"/cosmos.*.MsgSend"}}, types.ErrInvalidPattern},
		{"scope double star", s.authority, types.Params{Authority: s.authority, Scope: []string{"/cosmos.bank.*.*"}}, types.ErrInvalidPattern},
		{"scope digit-led segment", s.authority, types.Params{Authority: s.authority, Scope: []string{"/1cosmos.*"}}, types.ErrInvalidPattern},
		{"scope space", s.authority, types.Params{Authority: s.authority, Scope: []string{"/cosmos bank.*"}}, types.ErrInvalidPattern},
		{"scope double wildcard", s.authority, types.Params{Authority: s.authority, Scope: []string{"**"}}, types.ErrInvalidPattern},
		// sdk.ValidateAuthority compares strings, so an uppercase authority would
		// reject msgs signed with the usual lowercase address
		{"uppercase new authority", s.authority, types.Params{Authority: strings.ToUpper(s.authority)}, sdkerrors.ErrInvalidAddress},
		{"uppercase operator", s.authority, types.Params{Authority: s.authority, Operator: strings.ToUpper(s.operator)}, sdkerrors.ErrInvalidAddress},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.state(s.operator, []string{types.Wildcard}, uSend, "/cosmos.staking.*")
			before := s.params()
			ctx := s.freshEvents()

			_, err := s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{Authority: tc.signer, Params: tc.params})
			s.Require().ElementsMatch([]string{"/cosmos.staking.*", uSend}, s.list(), "UpdateParams must never change the disabled list")
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
				s.Require().Equal(before, s.params(), "rejected UpdateParams must not write params")
				s.Require().Empty(s.typedEvents(ctx))
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.params, s.params())
			s.Require().Equal([]proto.Message{&types.EventParamsUpdated{Authority: tc.signer, Params: tc.params}}, s.typedEvents(ctx))
		})
	}
}

// The current authority, read from params, can hand the role on; the
// consensus AuthorityParams override wins over params when set.
func (s *KeeperTestSuite) TestUpdateParamsAuthorityHandoffAndOverride() {
	newAuth := sdk.AccAddress("unit_new_authority__").String()
	_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.Params{Authority: newAuth}})
	s.Require().NoError(err)

	_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.Params{Authority: s.authority}})
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized, "the old authority lost the role")
	_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: newAuth, Params: types.Params{Authority: s.authority}})
	s.Require().NoError(err)

	override := sdk.AccAddress("unit_cons_authority_").String()
	ctx := s.freshEvents().WithConsensusParams(cmtproto.ConsensusParams{Authority: &cmtproto.AuthorityParams{Authority: override}})
	_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.Params{Authority: s.authority}})
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized, "params authority loses to the consensus override")
	_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{Authority: override, Params: types.Params{Authority: s.authority}})
	s.Require().NoError(err)
	// the event names the signer that passed the check, not the stored authority
	s.Require().Equal([]proto.Message{&types.EventParamsUpdated{Authority: override, Params: types.Params{Authority: s.authority}}}, s.typedEvents(ctx))

	// the consensus module accepts a non-canonical override; it still matches
	// a signer that uses the same string, as in every other module
	upper := strings.ToUpper(override)
	ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{Authority: &cmtproto.AuthorityParams{Authority: upper}})
	_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{Authority: upper, Params: types.Params{Authority: s.authority}})
	s.Require().NoError(err)
	s.state(s.operator, nil, uSend)
	_, err = s.msgServer.Reset(ctx, types.NewMsgReset(upper, uSend))
	s.Require().NoError(err)

	// an empty override string falls back to params
	ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{Authority: &cmtproto.AuthorityParams{Authority: ""}})
	_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.Params{Authority: s.authority}})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestTripTable() {
	all := []string{types.Wildcard}
	cases := []struct {
		name     string
		operator string
		scope    []string
		disabled []string
		sender   string
		patterns []string
		err      error
	}{
		{"exact", s.operator, all, nil, "", []string{uSend}, nil},
		{"prefix first segment", s.operator, all, nil, "", []string{"/cosmos.*"}, nil},
		{"prefix second segment", s.operator, all, nil, "", []string{"/cosmos.bank.*"}, nil},
		{"prefix third segment", s.operator, all, nil, "", []string{"/cosmos.bank.v1beta1.*"}, nil},
		{"wildcard", s.operator, all, nil, "", []string{types.Wildcard}, nil},
		{"several at once", s.operator, all, nil, "", []string{uSend, "/cosmos.staking.*", "/cosmos.authz.v1beta1.MsgGrant"}, nil},
		{"prefix over an already disabled exact", s.operator, all, []string{uSend}, "", []string{"/cosmos.bank.*"}, nil},
		{"scope exact covers itself", s.operator, []string{uSend}, nil, "", []string{uSend}, nil},
		{"scope prefix covers exact", s.operator, []string{"/cosmos.bank.*"}, nil, "", []string{uMultiSend}, nil},
		{"scope prefix covers deeper prefix", s.operator, []string{"/cosmos.*"}, nil, "", []string{"/cosmos.staking.v1beta1.*"}, nil},
		{"scope prefix covers itself", s.operator, []string{"/cosmos.staking.*"}, nil, "", []string{"/cosmos.staking.*"}, nil},
		{"second scope entry covers", s.operator, []string{uSend, "/cosmos.staking.*"}, nil, "", []string{uDelegate}, nil},
		{"operator equal to authority", s.authority, all, nil, s.authority, []string{uSend}, nil},
		{"authz exec", s.operator, all, nil, "", []string{uExec}, nil},
		{"authz prefix", s.operator, all, nil, "", []string{"/cosmos.authz.*"}, nil},
		// "*" and prefix entries do not block a wrapper, so an exact trip of
		// one under them, or beside them in one msg, changes something
		{"authz exec under a disabled wildcard", s.operator, all, []string{types.Wildcard}, "", []string{uExec}, nil},
		{"authz exec under a disabled authz prefix", s.operator, all, []string{"/cosmos.authz.*", "/cosmos.authz.v1beta1.*"}, "", []string{uExec}, nil},
		{"wildcard and authz exec in one msg", s.operator, all, nil, "", []string{types.Wildcard, uExec}, nil},
		{"authz prefix and authz exec in one msg", s.operator, all, nil, "", []string{"/cosmos.authz.*", uExec}, nil},
		{"authz prefix over a disabled authz exec", s.operator, all, []string{uExec}, "", []string{"/cosmos.authz.*"}, nil},
		{"authz exec in an authz prefix scope", s.operator, []string{"/cosmos.authz.*"}, nil, "", []string{uExec}, nil},
		{"operator-authority trips authz exec under a disabled wildcard", s.authority, []string{"/cosmos.authz.*"}, []string{types.Wildcard}, s.authority, []string{uExec}, nil},
		// a pattern not in the list is accepted under a broader disabled entry,
		// in or out of scope, and beside a broader pattern in the same msg, so
		// the operator can pin it before the broad entry is reset
		{"exact under an out-of-scope disabled prefix", s.operator, []string{uSend}, []string{"/cosmos.bank.*"}, "", []string{uSend}, nil},
		{"prefix under an out-of-scope disabled wildcard", s.operator, []string{"/cosmos.bank.*"}, []string{types.Wildcard}, "", []string{"/cosmos.bank.*"}, nil},
		{"exact under a disabled prefix", s.operator, all, []string{"/cosmos.bank.*"}, "", []string{uSend}, nil},
		{"prefix under a disabled prefix", s.operator, all, []string{"/cosmos.*"}, "", []string{"/cosmos.bank.*"}, nil},
		{"exact under disabled wildcard", s.operator, all, []string{types.Wildcard}, "", []string{uSend}, nil},
		{"exact under an in-scope prefix and an out-of-scope wildcard", s.operator, []string{"/cosmos.bank.*"}, []string{types.Wildcard, "/cosmos.bank.*"}, "", []string{uSend}, nil},
		{"authz grant under a disabled authz prefix", s.operator, all, []string{"/cosmos.authz.*"}, "", []string{"/cosmos.authz.v1beta1.MsgGrant"}, nil},
		{"prefix and an exact it covers in one msg", s.operator, all, nil, "", []string{"/cosmos.staking.*", uDelegate}, nil},
		{"exact before a prefix covering it in one msg", s.operator, all, nil, "", []string{uDelegate, "/cosmos.staking.*"}, nil},
		{"wildcard and an exact in one msg", s.operator, all, nil, "", []string{types.Wildcard, uSend}, nil},

		{"empty list", s.operator, all, nil, "", nil, types.ErrEmptyPatterns},
		{"duplicate", s.operator, all, nil, "", []string{uSend, uSend}, types.ErrDuplicatePattern},
		{"malformed no slash", s.operator, all, nil, "", []string{"cosmos.bank.v1beta1.MsgSend"}, types.ErrInvalidPattern},
		{"malformed trailing dot", s.operator, all, nil, "", []string{"/cosmos.bank."}, types.ErrInvalidPattern},
		{"malformed batch member", s.operator, all, nil, "", []string{uSend, "/cosmos..*"}, types.ErrInvalidPattern},
		{"empty sender", s.operator, all, nil, "-", []string{uSend}, sdkerrors.ErrInvalidAddress},
		{"malformed sender", s.operator, all, nil, "cosmos1xyz", []string{uSend}, sdkerrors.ErrInvalidAddress},
		{"uppercase operator", s.operator, all, nil, strings.ToUpper(s.operator), []string{uSend}, sdkerrors.ErrInvalidAddress},

		{"no operator configured", "", all, nil, s.operator, []string{uSend}, sdkerrors.ErrUnauthorized},
		{"no operator configured, authority", "", all, nil, s.authority, []string{uSend}, sdkerrors.ErrUnauthorized},
		{"stranger", s.operator, all, nil, s.stranger, []string{uSend}, sdkerrors.ErrUnauthorized},
		{"authority is not operator", s.operator, all, nil, s.authority, []string{uSend}, sdkerrors.ErrUnauthorized},

		{"empty scope", s.operator, nil, nil, "", []string{uSend}, types.ErrOutOfScope},
		{"empty scope wildcard", s.operator, nil, nil, "", []string{types.Wildcard}, types.ErrOutOfScope},
		{"prefix scope does not cover wildcard", s.operator, []string{"/cosmos.*"}, nil, "", []string{types.Wildcard}, types.ErrOutOfScope},
		{"prefix scope does not cover parent", s.operator, []string{"/cosmos.bank.*"}, nil, "", []string{"/cosmos.*"}, types.ErrOutOfScope},
		{"prefix scope does not cover sibling", s.operator, []string{"/cosmos.bank.*"}, nil, "", []string{uDelegate}, types.ErrOutOfScope},
		{"prefix scope stops at segment boundary", s.operator, []string{"/cosmos.staking.*"}, nil, "", []string{"/cosmos.stakingx.*"}, types.ErrOutOfScope},
		{"exact scope does not cover its package", s.operator, []string{uSend}, nil, "", []string{"/cosmos.bank.v1beta1.*"}, types.ErrOutOfScope},
		{"exact scope does not cover sibling", s.operator, []string{uSend}, nil, "", []string{uMultiSend}, types.ErrOutOfScope},
		{"out of scope batch member", s.operator, []string{"/cosmos.bank.*"}, nil, "", []string{uSend, uDelegate}, types.ErrOutOfScope},

		{"unregistered exact", s.operator, all, nil, "", []string{"/cosmos.bank.v1beta1.MsgSnd"}, types.ErrNoMatchingMsg},
		{"unregistered prefix", s.operator, all, nil, "", []string{"/nothing.here.*"}, types.ErrNoMatchingMsg},
		{"partial segment prefix", s.operator, all, nil, "", []string{"/cosmos.ban.*"}, types.ErrNoMatchingMsg},
		{"prefix past a msg name", s.operator, all, nil, "", []string{uSend + ".*"}, types.ErrNoMatchingMsg},
		{"protected safety reset", s.operator, all, nil, "", []string{uSafetyReset}, types.ErrNoMatchingMsg},
		{"protected safety trip", s.operator, all, nil, "", []string{uSafetyTrip}, types.ErrNoMatchingMsg},
		{"protected safety prefix", s.operator, all, nil, "", []string{"/cosmos.safety.*"}, types.ErrNoMatchingMsg},
		{"protected safety version prefix", s.operator, all, nil, "", []string{"/cosmos.safety.v1.*"}, types.ErrNoMatchingMsg},
		{"protected gov exact", s.operator, all, nil, "", []string{uVote}, types.ErrNoMatchingMsg},
		{"protected gov prefix", s.operator, all, nil, "", []string{"/cosmos.gov.*"}, types.ErrNoMatchingMsg},
		{"protected gov v1beta1 prefix", s.operator, all, nil, "", []string{"/cosmos.gov.v1beta1.*"}, types.ErrNoMatchingMsg},
		{"protected upgrade exact", s.operator, all, nil, "", []string{uUpgrade}, types.ErrNoMatchingMsg},
		{"protected upgrade prefix", s.operator, all, nil, "", []string{"/cosmos.upgrade.*"}, types.ErrNoMatchingMsg},
		{"protected consensus exact", s.operator, all, nil, "", []string{uConsensus}, types.ErrNoMatchingMsg},
		{"consensus prefix covers only protected msgs", s.operator, all, nil, "", []string{"/cosmos.consensus.*"}, types.ErrNoMatchingMsg},
		{"unregistered batch member", s.operator, all, nil, "", []string{uSend, "/does.not.Exist"}, types.ErrNoMatchingMsg},

		{"already disabled", s.operator, all, []string{uSend}, "", []string{uSend}, types.ErrAlreadyDisabled},
		{"already disabled prefix", s.operator, all, []string{"/cosmos.bank.*"}, "", []string{"/cosmos.bank.*"}, types.ErrAlreadyDisabled},
		{"already disabled wildcard", s.operator, all, []string{types.Wildcard}, "", []string{types.Wildcard}, types.ErrAlreadyDisabled},
		{"already disabled batch member", s.operator, all, []string{uSend}, "", []string{uMultiSend, uSend}, types.ErrAlreadyDisabled},
		{"authz exec already disabled", s.operator, all, []string{uExec, types.Wildcard}, "", []string{uExec}, types.ErrAlreadyDisabled},
		{"authz exec out of a bank scope", s.operator, []string{"/cosmos.bank.*"}, nil, "", []string{uExec}, types.ErrOutOfScope},
		{"already disabled out of scope cover", s.operator, []string{uSend}, []string{types.Wildcard, uSend}, "", []string{uSend}, types.ErrAlreadyDisabled},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.state(tc.operator, tc.scope, tc.disabled...)
			before := s.list()
			sender := tc.sender
			switch sender {
			case "":
				sender = s.operator
			case "-":
				sender = ""
			}
			ctx := s.freshEvents()

			_, err := s.msgServer.Trip(ctx, &types.MsgTrip{Operator: sender, Patterns: tc.patterns})
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
				s.requireRegisteredError(err)
				s.Require().Equal(before, s.list(), "a rejected trip must not write state")
				s.Require().Empty(s.typedEvents(ctx))
				return
			}
			s.Require().NoError(err)
			s.Require().ElementsMatch(append(append([]string{}, tc.disabled...), tc.patterns...), s.list())
			s.Require().Equal([]proto.Message{&types.EventTrip{Operator: sender, Patterns: tc.patterns}}, s.typedEvents(ctx))
			for _, p := range tc.patterns {
				if p != types.Wildcard && !types.IsPrefixPattern(p) {
					s.Require().False(s.allowed(p), "%s tripped but still allowed", p)
				}
			}
		})
	}
}

// After a successful trip, IsAllowed blocks every registered unprotected msg
// the pattern covers, except a wrapper under a broader pattern, and nothing
// else.
func (s *KeeperTestSuite) TestTripEffectOnRegisteredMsgs() {
	registered := s.registry.ListImplementations(sdk.MsgInterfaceProtoName)
	for _, p := range []string{uSend, "/cosmos.bank.*", "/cosmos.staking.v1beta1.*", "/cosmos.*", types.Wildcard, "/cosmos.authz.*", uExec} {
		s.Run(p, func() {
			s.state(s.operator, []string{types.Wildcard})
			_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, p))
			s.Require().NoError(err)
			for _, url := range registered {
				want := types.IsProtected(url) || !types.Covers(p, url) || (types.IsWrapper(url) && p != url)
				s.Require().Equal(want, s.allowed(url), "%s after tripping %s", url, p)
			}
		})
	}
}

func (s *KeeperTestSuite) TestResetTable() {
	disabled := []string{uSend, "/cosmos.staking.*", types.Wildcard, uSafetyReset, uMultiSend, uExec}
	bank := []string{"/cosmos.bank.*"}
	cases := []struct {
		name     string
		operator string
		scope    []string
		sender   string
		patterns []string
		err      error
	}{
		{"authority resets exact", s.operator, bank, s.authority, []string{uSend}, nil},
		{"authority resets out of operator scope", s.operator, bank, s.authority, []string{"/cosmos.staking.*"}, nil},
		{"authority resets wildcard", s.operator, bank, s.authority, []string{types.Wildcard}, nil},
		{"authority resets protected entry", s.operator, bank, s.authority, []string{uSafetyReset}, nil},
		{"authority resets with no operator", "", nil, s.authority, []string{uSend, types.Wildcard}, nil},
		{"authority resets with empty scope", s.operator, nil, s.authority, []string{"/cosmos.staking.*"}, nil},
		{"operator resets in scope", s.operator, bank, s.operator, []string{uSend}, nil},
		{"operator resets several in scope", s.operator, bank, s.operator, []string{uSend, uMultiSend}, nil},
		{"operator with wildcard scope resets wildcard", s.operator, []string{types.Wildcard}, s.operator, []string{types.Wildcard}, nil},
		{"operator equal to authority resets anything", s.authority, nil, s.authority, []string{"/cosmos.staking.*"}, nil},
		{"authority resets a wrapper", s.operator, bank, s.authority, []string{uExec}, nil},
		{"operator with wildcard scope resets a wrapper", s.operator, []string{types.Wildcard}, s.operator, []string{uExec}, nil},
		{"operator with authz prefix scope resets a wrapper", s.operator, []string{"/cosmos.authz.*"}, s.operator, []string{uExec}, nil},

		{"operator out of scope", s.operator, bank, s.operator, []string{"/cosmos.staking.*"}, types.ErrOutOfScope},
		{"operator out of scope wildcard", s.operator, bank, s.operator, []string{types.Wildcard}, types.ErrOutOfScope},
		{"operator empty scope", s.operator, nil, s.operator, []string{uSend}, types.ErrOutOfScope},
		{"operator out of scope batch member", s.operator, bank, s.operator, []string{uSend, "/cosmos.staking.*"}, types.ErrOutOfScope},
		{"operator out of scope protected entry", s.operator, bank, s.operator, []string{uSafetyReset}, types.ErrOutOfScope},
		{"operator out of scope wrapper", s.operator, bank, s.operator, []string{uExec}, types.ErrOutOfScope},
		{"authz prefix is not an exact match for a wrapper", s.operator, []string{types.Wildcard}, s.operator, []string{"/cosmos.authz.*"}, types.ErrNotDisabled},
		{"stranger", s.operator, bank, s.stranger, []string{uSend}, sdkerrors.ErrUnauthorized},
		{"no operator configured", "", []string{types.Wildcard}, s.operator, []string{uSend}, sdkerrors.ErrUnauthorized},

		{"not disabled", s.operator, bank, s.authority, []string{uDelegate}, types.ErrNotDisabled},
		{"covering prefix is not an exact match", s.operator, bank, s.authority, []string{"/cosmos.bank.*"}, types.ErrNotDisabled},
		{"covered exact is not an exact match", s.operator, bank, s.authority, []string{uDelegate}, types.ErrNotDisabled},
		{"deeper prefix is not an exact match", s.operator, bank, s.authority, []string{"/cosmos.staking.v1beta1.*"}, types.ErrNotDisabled},
		{"operator not disabled in scope", s.operator, bank, s.operator, []string{"/cosmos.bank.v1beta1.*"}, types.ErrNotDisabled},
		{"not disabled batch member", s.operator, bank, s.authority, []string{uSend, uDelegate}, types.ErrNotDisabled},

		{"empty list", s.operator, bank, s.authority, nil, types.ErrEmptyPatterns},
		{"duplicate", s.operator, bank, s.authority, []string{uSend, uSend}, types.ErrDuplicatePattern},
		{"malformed", s.operator, bank, s.authority, []string{"bank"}, types.ErrInvalidPattern},
		{"empty sender", s.operator, bank, "", []string{uSend}, sdkerrors.ErrInvalidAddress},
		{"malformed sender", s.operator, bank, "cosmos1xyz", []string{uSend}, sdkerrors.ErrInvalidAddress},
		// roles match by exact string, as sdk.ValidateAuthority does, so an
		// uppercase sender is neither role and does not reach the scope check;
		// the error names the canonical form instead of calling it a stranger
		{"uppercase operator", s.operator, bank, strings.ToUpper(s.operator), []string{uSend}, sdkerrors.ErrInvalidAddress},
		{"uppercase authority", s.operator, bank, strings.ToUpper(s.authority), []string{uSend}, sdkerrors.ErrInvalidAddress},
		{"uppercase authority that is also operator", s.authority, bank, strings.ToUpper(s.authority), []string{"/cosmos.staking.*"}, sdkerrors.ErrInvalidAddress},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.state(tc.operator, tc.scope, disabled...)
			before := s.list()
			ctx := s.freshEvents()

			_, err := s.msgServer.Reset(ctx, &types.MsgReset{Sender: tc.sender, Patterns: tc.patterns})
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
				s.requireRegisteredError(err)
				s.Require().Equal(before, s.list(), "a rejected reset must not write state")
				s.Require().Empty(s.typedEvents(ctx))
				return
			}
			s.Require().NoError(err)
			var want []string
			for _, p := range disabled {
				removed := false
				for _, r := range tc.patterns {
					removed = removed || r == p
				}
				if !removed {
					want = append(want, p)
				}
			}
			s.Require().ElementsMatch(want, s.list())
			s.Require().Equal([]proto.Message{&types.EventReset{Sender: tc.sender, Patterns: tc.patterns}}, s.typedEvents(ctx))
		})
	}
}

// With the consensus override set, the params authority is no longer the
// authority for reset, and falls back to the operator path.
func (s *KeeperTestSuite) TestResetConsensusOverride() {
	override := sdk.AccAddress("unit_cons_authority_").String()
	ctx := s.ctx.WithConsensusParams(cmtproto.ConsensusParams{Authority: &cmtproto.AuthorityParams{Authority: override}})
	s.state(s.operator, []string{"/cosmos.bank.*"}, uSend, uDelegate)

	_, err := s.msgServer.Reset(ctx, types.NewMsgReset(s.authority, uDelegate))
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
	_, err = s.msgServer.Reset(ctx, types.NewMsgReset(override, uDelegate))
	s.Require().NoError(err)
	_, err = s.msgServer.Reset(ctx, types.NewMsgReset(s.operator, uSend))
	s.Require().NoError(err)
	s.Require().Empty(s.list())
}

// Trip then reset restores the exact prior state, for each pattern form.
func (s *KeeperTestSuite) TestTripResetRoundTrip() {
	for _, p := range []string{uSend, "/cosmos.*", "/cosmos.bank.*", "/cosmos.bank.v1beta1.*", types.Wildcard} {
		s.Run(p, func() {
			s.state(s.operator, []string{types.Wildcard})
			s.Require().True(s.allowed(uSend))
			_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, p))
			s.Require().NoError(err)
			s.Require().False(s.allowed(uSend))
			s.Require().True(s.allowed(uSafetyReset))
			_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, p))
			s.Require().NoError(err)
			s.Require().True(s.allowed(uSend))
			s.Require().Empty(s.list())
		})
	}
}

// After the authority narrows the operator's scope below a broad trip, the
// operator pins the msg it still covers before the authority resets the
// broad entry, so the msg never reopens.
func (s *KeeperTestSuite) TestNarrowScopeUnderBroadTrip() {
	s.state(s.operator, []string{"/cosmos.*"})
	_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, "/cosmos.bank.*"))
	s.Require().NoError(err)

	_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params:    types.Params{Authority: s.authority, Operator: s.operator, Scope: []string{uSend}},
	})
	s.Require().NoError(err)

	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uSend))
	s.Require().NoError(err)
	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, "/cosmos.bank.*"))
	s.Require().ErrorIs(err, types.ErrOutOfScope)
	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.authority, "/cosmos.bank.*"))
	s.Require().NoError(err)

	s.Require().False(s.allowed(uSend))
	s.Require().True(s.allowed(uMultiSend))
	s.Require().Equal([]string{uSend}, s.list())
}

// The operator cuts off a wrapper while "*" stands by tripping its exact URL,
// and the wrapper stays blocked after "*" is reset.
func (s *KeeperTestSuite) TestTripWrapperUnderWildcard() {
	s.state(s.operator, []string{types.Wildcard}, types.Wildcard)
	s.Require().True(s.allowed(uExec))
	s.Require().False(s.allowed(uSend))

	_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uExec))
	s.Require().NoError(err)
	s.Require().False(s.allowed(uExec))

	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, types.Wildcard))
	s.Require().NoError(err)
	s.Require().False(s.allowed(uExec))
	s.Require().True(s.allowed(uSend))
	s.Require().Equal([]string{uExec}, s.list())

	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, uExec))
	s.Require().NoError(err)
	s.Require().True(s.allowed(uExec))
}

// A prefix that covers only a wrapper among the registered msgs blocks
// nothing, so a trip of it fails like any trip that matches no msg. An
// exact trip of the wrapper still succeeds.
func (s *KeeperTestSuite) TestTripPrefixCoveringOnlyAWrapper() {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	encCfg.InterfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &authz.MsgExec{})
	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(s.key), addresscodec.NewBech32Codec("cosmos"), encCfg.InterfaceRegistry, s.authority)
	srv := keeper.NewMsgServerForTest(k)
	s.state(s.operator, []string{types.Wildcard})

	for _, p := range []string{"/cosmos.authz.*", "/cosmos.authz.v1beta1.*", "/cosmos.*"} {
		_, err := srv.Trip(s.ctx, types.NewMsgTrip(s.operator, p))
		s.Require().ErrorIs(err, types.ErrNoMatchingMsg, p)
	}
	s.Require().Empty(s.list())

	_, err := srv.Trip(s.ctx, types.NewMsgTrip(s.operator, uExec))
	s.Require().NoError(err)
	s.Require().Equal([]string{uExec}, s.list())
}

// A trip needs only patterns that are not already in the disabled list.
// The operator pins a msg under a disabled "*" it could reset itself, so the
// msg stays blocked after "*" is reset, and a prefix and an exact pattern it
// covers can be tripped in one msg. An operator that is also the authority,
// directly or through the consensus override, pins the same way.
func (s *KeeperTestSuite) TestTripPinsUnderCoveringEntries() {
	s.state(s.operator, []string{types.Wildcard}, types.Wildcard)
	_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uSend))
	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{types.Wildcard, uSend}, s.list())
	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, types.Wildcard))
	s.Require().NoError(err)
	s.Require().False(s.allowed(uSend))
	s.Require().True(s.allowed(uMultiSend))

	s.state(s.operator, []string{types.Wildcard})
	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, "/cosmos.bank.*", uSend))
	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{"/cosmos.bank.*", uSend}, s.list())
	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, "/cosmos.bank.*"))
	s.Require().NoError(err)
	s.Require().False(s.allowed(uSend))
	s.Require().True(s.allowed(uMultiSend))

	s.state(s.authority, []string{"/cosmos.bank.*"}, types.Wildcard)
	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.authority, uSend))
	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{types.Wildcard, uSend}, s.list())

	override := s.ctx.WithConsensusParams(cmtproto.ConsensusParams{Authority: &cmtproto.AuthorityParams{Authority: s.operator}})
	s.state(s.operator, []string{"/cosmos.bank.*"}, types.Wildcard)
	_, err = s.msgServer.Trip(override, types.NewMsgTrip(s.operator, uSend))
	s.Require().NoError(err)
	s.Require().ElementsMatch([]string{types.Wildcard, uSend}, s.list())
}

func (s *KeeperTestSuite) TestQueryParams() {
	res, err := s.queryClient.Params(s.ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.params(), res.Params)

	want := types.Params{Authority: s.authority, Operator: s.other, Scope: []string{uSend, "/cosmos.staking.*"}}
	_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: want})
	s.Require().NoError(err)
	res, err = s.queryClient.Params(s.ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(want, res.Params)
}

func (s *KeeperTestSuite) TestQueryNilRequests() {
	_, err := s.queryServer.Params(s.ctx, nil)
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
	_, err = s.queryServer.DisabledList(s.ctx, nil)
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
	_, err = s.queryServer.IsAllowed(s.ctx, nil)
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
	_, err = s.queryServer.IsAllowed(s.ctx, &types.QueryIsAllowedRequest{})
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
}

// The IsAllowed query takes an arbitrary string from an unauthenticated
// caller. It must reject anything that is not a well-formed msg type URL, and
// do so before any work that grows with the input.
func (s *KeeperTestSuite) TestQueryIsAllowedRejectsMalformed() {
	s.state(s.operator, []string{types.Wildcard}, "/cosmos.bank.*")
	for _, u := range []string{
		"cosmos.bank.v1beta1.MsgSend", // no leading slash: would never match a prefix
		"/cosmos.bank.*",
		types.Wildcard,
		"/",
		"/cosmos..MsgSend",
		"/" + strings.Repeat("a.", types.MaxPatternLength) + "MsgX",
	} {
		_, err := s.queryClient.IsAllowed(s.ctx, &types.QueryIsAllowedRequest{MsgTypeUrl: u})
		s.Require().Equal(codes.InvalidArgument, status.Code(err), "%.40q: %v", u, err)
	}

	// a long dotted input uses memory linear in its size, not quadratic
	long := "/" + strings.Repeat("a.", 8_000) + "MsgX"
	var before, after goruntime.MemStats
	goruntime.GC()
	goruntime.ReadMemStats(&before)
	_, err := s.queryClient.IsAllowed(s.ctx, &types.QueryIsAllowedRequest{MsgTypeUrl: long})
	goruntime.ReadMemStats(&after)
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
	s.Require().Less(after.TotalAlloc-before.TotalAlloc, uint64(1<<20), "allocated bytes for a %d byte query", len(long))
}

func (s *KeeperTestSuite) TestQueryDisabledList() {
	res, err := s.queryClient.DisabledList(s.ctx, &types.QueryDisabledListRequest{Pagination: &querytypes.PageRequest{CountTotal: true}})
	s.Require().NoError(err)
	s.Require().Empty(res.DisabledPatterns)
	s.Require().Equal(uint64(0), res.Pagination.Total)

	all := []string{types.Wildcard, "/cosmos.bank.*", uSend, "/cosmos.staking.*", uDelegate}
	s.state(s.operator, []string{types.Wildcard}, all...)
	// the store orders keys bytewise, and "*" sorts before "/" and before letters
	s.Require().Equal([]string{types.Wildcard, "/cosmos.bank.*", uSend, "/cosmos.staking.*", uDelegate}, s.list())

	res, err = s.queryClient.DisabledList(s.ctx, &types.QueryDisabledListRequest{})
	s.Require().NoError(err)
	s.Require().ElementsMatch(all, res.DisabledPatterns)

	cases := []struct {
		name string
		page *querytypes.PageRequest
		want []string
	}{
		{"limit", &querytypes.PageRequest{Limit: 2}, s.list()[:2]},
		{"offset", &querytypes.PageRequest{Offset: 3, Limit: 10}, s.list()[3:]},
		{"reverse", &querytypes.PageRequest{Limit: 2, Reverse: true}, []string{s.list()[4], s.list()[3]}},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			res, err := s.queryClient.DisabledList(s.ctx, &types.QueryDisabledListRequest{Pagination: tc.page})
			s.Require().NoError(err)
			s.Require().Equal(tc.want, res.DisabledPatterns)
		})
	}

	// an offset past the end returns no patterns and still counts all of them
	res, err = s.queryClient.DisabledList(s.ctx, &types.QueryDisabledListRequest{Pagination: &querytypes.PageRequest{Offset: 7, CountTotal: true}})
	s.Require().NoError(err)
	s.Require().Empty(res.DisabledPatterns)
	s.Require().Equal(uint64(len(all)), res.Pagination.Total)

	_, err = s.queryClient.DisabledList(s.ctx, &types.QueryDisabledListRequest{Pagination: &querytypes.PageRequest{Key: []byte("x"), Offset: 1}})
	s.Require().Equal(codes.InvalidArgument, status.Code(err), "key and offset together are invalid: %v", err)
}

// Every registered x/safety error lives in the safety codespace with a unique code.
func (s *KeeperTestSuite) TestRegisteredErrors() {
	errs := []*errorsmod.Error{
		types.ErrInvalidPattern, types.ErrEmptyPatterns, types.ErrDuplicatePattern,
		types.ErrInvalidParams, types.ErrOutOfScope, types.ErrNoMatchingMsg,
		types.ErrAlreadyDisabled, types.ErrNotDisabled, types.ErrMsgDisabled,
	}
	seen := map[uint32]bool{}
	for _, e := range errs {
		s.Require().Equal(types.ModuleName, e.Codespace(), e.Error())
		s.Require().Greater(e.ABCICode(), uint32(1), "code 1 is reserved for internal errors")
		s.Require().False(seen[e.ABCICode()], "duplicate code %d", e.ABCICode())
		seen[e.ABCICode()] = true

		wrapped := e.Wrap("context")
		s.Require().ErrorIs(wrapped, e)
		s.requireRegisteredError(wrapped)
	}
}

// NewKeeper rejects any default authority that the default genesis would then
// fail to validate, so a wiring mistake surfaces at construction, not at init.
func (s *KeeperTestSuite) TestNewKeeperRejectsInvalidDefaultAuthority() {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	for _, authority := range []string{"", "not-an-address", strings.ToUpper(s.authority)} {
		s.Require().Panics(func() {
			keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(s.key), addresscodec.NewBech32Codec("cosmos"), encCfg.InterfaceRegistry, authority)
		}, "%q", authority)
	}
}

func (s *KeeperTestSuite) TestKeeperIsCircuitBreaker() {
	var cb baseapp.CircuitBreaker = s.keeper
	s.state(s.operator, []string{types.Wildcard}, uSend)
	ok, err := cb.IsAllowed(s.ctx, uSend)
	s.Require().ErrorIs(err, types.ErrMsgDisabled, "IsAllowed returns the registered error the router returns for a blocked msg")
	s.Require().False(ok)
}

func (s *KeeperTestSuite) TestDefaultGenesis() {
	gs := types.DefaultGenesisState(s.keeper.DefaultAuthority())
	s.Require().Equal(s.authority, gs.Params.Authority, "default authority is the gov module account")
	s.Require().Empty(gs.Params.Operator, "default operator is empty")
	s.Require().Empty(gs.DisabledPatterns)
	s.Require().NoError(gs.Validate(s.keeper.AddressCodec()))
}

// The default genesis scope is empty, so the operator can do
// nothing until the authority grants scope.
func (s *KeeperTestSuite) TestDefaultGenesisScopeEmpty() {
	gs := types.DefaultGenesisState(s.authority)
	s.Require().Empty(gs.Params.Scope, "default scope must be empty, got %v", gs.Params.Scope)
}

func (s *KeeperTestSuite) TestInitGenesisTable() {
	cases := []struct {
		name string
		gs   *types.GenesisState
		err  error
	}{
		{"default", types.NewGenesisState(types.Params{Authority: s.authority}, nil), nil},
		{"every form disabled", types.NewGenesisState(types.Params{Authority: s.authority, Operator: s.operator, Scope: []string{"/cosmos.*"}},
			[]string{types.Wildcard, "/cosmos.*", uSend}), nil},
		{"protected entries are accepted and inert", types.NewGenesisState(types.Params{Authority: s.authority},
			[]string{uSafetyReset, "/cosmos.gov.*", uConsensus}), nil},
		{"empty authority", types.NewGenesisState(types.Params{}, nil), types.ErrInvalidParams},
		{"malformed authority", types.NewGenesisState(types.Params{Authority: "x"}, nil), sdkerrors.ErrInvalidAddress},
		{"malformed operator", types.NewGenesisState(types.Params{Authority: s.authority, Operator: "x"}, nil), sdkerrors.ErrInvalidAddress},
		{"malformed scope", types.NewGenesisState(types.Params{Authority: s.authority, Scope: []string{"bank"}}, nil), types.ErrInvalidPattern},
		{"duplicate scope", types.NewGenesisState(types.Params{Authority: s.authority, Scope: []string{"*", "*"}}, nil), types.ErrDuplicatePattern},
		{"malformed disabled", types.NewGenesisState(types.Params{Authority: s.authority}, []string{"cosmos.bank"}), types.ErrInvalidPattern},
		{"empty disabled entry", types.NewGenesisState(types.Params{Authority: s.authority}, []string{""}), types.ErrInvalidPattern},
		{"duplicate disabled", types.NewGenesisState(types.Params{Authority: s.authority}, []string{uSend, uSend}), types.ErrDuplicatePattern},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.Require().NoError(s.keeper.Params.Remove(s.ctx))
			err := s.keeper.InitGenesis(s.ctx, tc.gs)
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
				has, herr := s.keeper.Params.Has(s.ctx)
				s.Require().NoError(herr)
				s.Require().False(has, "a rejected genesis must not write params")
				s.Require().Empty(s.list())
				return
			}
			s.Require().NoError(err)
			out, err := s.keeper.ExportGenesis(s.ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.gs.Params, out.Params)
			s.Require().ElementsMatch(tc.gs.DisabledPatterns, out.DisabledPatterns)
			for _, u := range []string{uSafetyReset, uVote, uConsensus, uUpgrade} {
				s.Require().True(s.allowed(u), "%s must stay allowed whatever genesis disables", u)
			}
		})
	}
}

var errStore = errors.New("store unavailable")

// failingStoreService opens stores whose Has always fails.
type failingStoreService struct{ inner corestore.KVStoreService }

func (f failingStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return failingStore{f.inner.OpenKVStore(ctx)}
}

type failingStore struct{ corestore.KVStore }

func (failingStore) Has([]byte) (bool, error) { return false, errStore }

// A store error blocks the msg: the breaker fails closed.
func (s *KeeperTestSuite) TestIsAllowedFailsClosedOnStoreError() {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	k := keeper.NewKeeper(encCfg.Codec, failingStoreService{runtime.NewKVStoreService(s.key)}, addresscodec.NewBech32Codec("cosmos"), s.registry, s.authority)

	ok, err := k.IsAllowed(s.ctx, uSend)
	s.Require().ErrorIs(err, errStore)
	s.Require().False(ok)

	_, err = keeper.NewQueryServer(k).IsAllowed(s.ctx, &types.QueryIsAllowedRequest{MsgTypeUrl: uSend})
	s.Require().Equal(codes.Internal, status.Code(err))

	// a protected msg needs no lookup and stays allowed
	ok, err = k.IsAllowed(s.ctx, uSafetyReset)
	s.Require().NoError(err)
	s.Require().True(ok)
}

// SetParams writes valid params and refuses invalid ones, leaving the stored
// params as they were.
func (s *KeeperTestSuite) TestSetParams() {
	want := types.Params{Authority: s.authority, Operator: s.other, Scope: []string{"/cosmos.bank.*"}}
	s.Require().NoError(s.keeper.SetParams(s.ctx, want))
	got, err := s.keeper.Params.Get(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(want, got)

	for _, bad := range []types.Params{
		{Authority: ""},
		{Authority: s.authority, Operator: "not-an-address"},
		{Authority: s.authority, Scope: []string{"/cosmos.bank.*", "/cosmos.bank.*"}},
		{Authority: s.authority, Scope: []string{"cosmos.bank"}},
	} {
		s.Require().Error(s.keeper.SetParams(s.ctx, bad), "%+v", bad)
		got, err := s.keeper.Params.Get(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(want, got)
	}
}

// The IsAllowed query names the disabled entry that blocks a msg, so its JSON
// output says why next to allowed: false.
func (s *KeeperTestSuite) TestQueryIsAllowedBlockedBy() {
	s.state(s.operator, []string{types.Wildcard}, "/cosmos.bank.*")

	res, err := s.queryClient.IsAllowed(s.ctx, &types.QueryIsAllowedRequest{MsgTypeUrl: uSend})
	s.Require().NoError(err)
	s.Require().False(res.Allowed)
	s.Require().Equal("/cosmos.bank.*", res.BlockedBy)
	bz, err := codec.ProtoMarshalJSON(res, nil)
	s.Require().NoError(err)
	s.Require().JSONEq(`{"allowed":false,"blocked_by":"/cosmos.bank.*"}`, string(bz))

	res, err = s.queryClient.IsAllowed(s.ctx, &types.QueryIsAllowedRequest{MsgTypeUrl: uDelegate})
	s.Require().NoError(err)
	s.Require().True(res.Allowed)
	s.Require().Empty(res.BlockedBy)
}

func (s *KeeperTestSuite) matching(pattern string, page *querytypes.PageRequest) (*types.QueryMatchingMsgsResponse, error) {
	return s.queryClient.MatchingMsgs(s.ctx, &types.QueryMatchingMsgsRequest{Pattern: pattern, Pagination: page})
}

// MatchingMsgs lists what a trip of a pattern would block: the registered
// msgs IsAllowed would refuse, with protected msgs left out and wrappers only
// for an exact pattern.
func (s *KeeperTestSuite) TestQueryMatchingMsgs() {
	all := s.registry.ListImplementations(sdk.MsgInterfaceProtoName)
	for _, pattern := range []string{types.Wildcard, "/cosmos.*", "/cosmos.bank.*", "/cosmos.authz.v1beta1.*", uSend, uExec, uVote, "/cosmos.gov.*", "/other.*"} {
		s.Run(pattern, func() {
			res, err := s.matching(pattern, nil)
			s.Require().NoError(err)

			// the list is exactly what IsAllowed refuses once the pattern
			// is tripped
			s.state(s.operator, []string{types.Wildcard}, pattern)
			var want []string
			for _, u := range all {
				if !s.allowed(u) {
					want = append(want, u)
				}
			}
			slices.Sort(want)
			s.Require().Equal(want, res.MsgTypeUrls)
		})
	}

	res, err := s.matching("/cosmos.bank.*", nil)
	s.Require().NoError(err)
	s.Require().Contains(res.MsgTypeUrls, uSend)
	s.Require().Contains(res.MsgTypeUrls, uMultiSend)
	s.Require().NotContains(res.MsgTypeUrls, uDelegate)

	res, err = s.matching(types.Wildcard, nil)
	s.Require().NoError(err)
	s.Require().Contains(res.MsgTypeUrls, uSend)
	for _, u := range []string{uExec, uVote, uSubmitProp, uSafetyReset, uUpgrade, uConsensus} {
		s.Require().NotContains(res.MsgTypeUrls, u)
	}

	res, err = s.matching(uExec, nil)
	s.Require().NoError(err)
	s.Require().Equal([]string{uExec}, res.MsgTypeUrls)

	res, err = s.matching(uVote, nil)
	s.Require().NoError(err)
	s.Require().Empty(res.MsgTypeUrls)

	for _, bad := range []string{"", "cosmos.bank", "/cosmos.bank.**", "/cosmos.*.MsgSend"} {
		_, err := s.matching(bad, nil)
		s.Require().Equal(codes.InvalidArgument, status.Code(err), bad)
	}
	_, err = s.queryServer.MatchingMsgs(s.ctx, nil)
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
}

// MatchingMsgs pages through the sorted list by offset or by key.
func (s *KeeperTestSuite) TestQueryMatchingMsgsPagination() {
	full, err := s.matching(types.Wildcard, nil)
	s.Require().NoError(err)
	want := full.MsgTypeUrls
	s.Require().Greater(len(want), 4)
	s.Require().True(slices.IsSorted(want))

	// by key, two at a time
	var got []string
	var key []byte
	for {
		res, err := s.matching(types.Wildcard, &querytypes.PageRequest{Key: key, Limit: 2})
		s.Require().NoError(err)
		s.Require().LessOrEqual(len(res.MsgTypeUrls), 2)
		got = append(got, res.MsgTypeUrls...)
		key = res.Pagination.NextKey
		if key == nil {
			break
		}
	}
	s.Require().Equal(want, got)

	// by offset, with the total
	res, err := s.matching(types.Wildcard, &querytypes.PageRequest{Offset: 1, Limit: 2, CountTotal: true})
	s.Require().NoError(err)
	s.Require().Equal(want[1:3], res.MsgTypeUrls)
	s.Require().Equal(uint64(len(want)), res.Pagination.Total)
	s.Require().Equal([]byte(want[3]), res.Pagination.NextKey)

	// an offset past the end returns nothing and the real total
	res, err = s.matching(types.Wildcard, &querytypes.PageRequest{Offset: uint64(len(want) + 5), CountTotal: true})
	s.Require().NoError(err)
	s.Require().Empty(res.MsgTypeUrls)
	s.Require().Nil(res.Pagination.NextKey)
	s.Require().Equal(uint64(len(want)), res.Pagination.Total)

	// a limit near the uint64 max with an offset returns the rest
	res, err = s.matching(types.Wildcard, &querytypes.PageRequest{Offset: 1, Limit: math.MaxUint64})
	s.Require().NoError(err)
	s.Require().Equal(want[1:], res.MsgTypeUrls)
	s.Require().Nil(res.Pagination.NextKey)

	// reverse
	res, err = s.matching(types.Wildcard, &querytypes.PageRequest{Limit: 2, Reverse: true})
	s.Require().NoError(err)
	s.Require().Equal([]string{want[len(want)-1], want[len(want)-2]}, res.MsgTypeUrls)

	// key and offset together are refused
	_, err = s.matching(types.Wildcard, &querytypes.PageRequest{Key: []byte(want[1]), Offset: 1})
	s.Require().Equal(codes.InvalidArgument, status.Code(err))
}

// A refusal logs one debug line naming the type URL and the blocking entry,
// and charges no gas or emits no event. An allowed msg logs nothing.
func (s *KeeperTestSuite) TestRefusalLogged() {
	s.state(s.operator, []string{types.Wildcard}, "/cosmos.bank.*")
	var buf bytes.Buffer
	ctx := s.freshEvents().
		WithLogger(log.NewLogger(&buf, log.OutputJSONOption(), log.LevelOption(zerolog.DebugLevel))).
		WithGasMeter(storetypes.NewGasMeter(1_000_000))

	_, err := s.keeper.IsAllowed(ctx, uDelegate)
	s.Require().NoError(err)
	s.Require().Empty(buf.String())

	_, err = s.keeper.IsAllowed(ctx, uSend)
	s.Require().ErrorIs(err, types.ErrMsgDisabled)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	s.Require().Len(lines, 1, buf.String())
	var entry map[string]any
	s.Require().NoError(json.Unmarshal([]byte(lines[0]), &entry))
	s.Require().Equal("debug", entry["level"])
	s.Require().Equal("x/"+types.ModuleName, entry["module"])
	s.Require().Equal(uSend, entry["type_url"])
	s.Require().Equal("/cosmos.bank.*", entry["pattern"])
	s.Require().Zero(ctx.GasMeter().GasConsumed())
	s.Require().Empty(ctx.EventManager().Events())
}

// Regression tests for the x/circuit behaviors that x/safety changes. The
// README section "Changes from x/circuit" lists them.

// x/circuit let an empty pattern list trip or reset successfully and do nothing.
func (s *KeeperTestSuite) TestEmptyPatternListRejected() {
	s.state(s.operator, []string{types.Wildcard}, uSend)
	for _, run := range []func() error{
		func() error { _, err := s.msgServer.Trip(s.ctx, &types.MsgTrip{Operator: s.operator}); return err },
		func() error {
			_, err := s.msgServer.Trip(s.ctx, &types.MsgTrip{Operator: s.operator, Patterns: []string{}})
			return err
		},
		func() error { _, err := s.msgServer.Reset(s.ctx, &types.MsgReset{Sender: s.authority}); return err },
		func() error {
			_, err := s.msgServer.Reset(s.ctx, &types.MsgReset{Sender: s.operator, Patterns: []string{}})
			return err
		},
	} {
		s.Require().ErrorIs(run(), types.ErrEmptyPatterns)
	}
	s.Require().Equal([]string{uSend}, s.list())
}

// x/circuit let any string, including typos and nonsense, be tripped.
func (s *KeeperTestSuite) TestOnlyRegisteredMsgsTrip() {
	cases := map[string]error{
		"/cosmos.bank.v1beta1.MsgSendd": types.ErrNoMatchingMsg,
		"/cosmos.bank.v1.*":             types.ErrNoMatchingMsg,
		"/foo.*":                        types.ErrNoMatchingMsg,
		"/cosmos.bank":                  types.ErrNoMatchingMsg,
		"hello":                         types.ErrInvalidPattern,
		"/cosmos.bank.v1beta1.MsgSend,/cosmos.bank.v1beta1.MsgMultiSend": types.ErrInvalidPattern,
		" /cosmos.bank.v1beta1.MsgSend":                                  types.ErrInvalidPattern,
		"/cosmos.bank.v1beta1.MsgSend ":                                  types.ErrInvalidPattern,
	}
	for p, want := range cases {
		_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, p))
		s.Require().ErrorIs(err, want, p)
	}
	s.Require().Empty(s.list())

	_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uSend))
	s.Require().NoError(err)
}

// x/circuit let the reset msg itself be tripped, locking out recovery.
func (s *KeeperTestSuite) TestNoResetLockout() {
	for _, p := range []string{uSafetyReset, "/cosmos.safety.*", "/cosmos.safety.v1.*", uSafetyParams, uVote, "/cosmos.gov.*", uUpgrade, uConsensus} {
		_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, p))
		s.Require().ErrorIs(err, types.ErrNoMatchingMsg, p)
	}

	// the broadest trips available still leave recovery running
	_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, "/cosmos.*"))
	s.Require().NoError(err)
	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, types.Wildcard))
	s.Require().NoError(err)
	for _, u := range []string{uSafetyReset, uSafetyParams, uSafetyTrip, uVote, uSubmitProp, uUpgrade, uConsensus} {
		s.Require().True(s.allowed(u), u)
	}

	// and a protected entry forced in through state is inert
	s.Require().NoError(s.keeper.DisabledList.Set(s.ctx, uSafetyReset))
	s.Require().True(s.allowed(uSafetyReset))

	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.authority, types.Wildcard, "/cosmos.*", uSafetyReset))
	s.Require().NoError(err)
	s.Require().Empty(s.list())
}

// x/circuit had a permission levels table and super admins who could
// delegate. x/safety has one operator, set only by the authority through params.
func (s *KeeperTestSuite) TestSingleOperator() {
	var safetyMsgs []string
	for _, u := range s.registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		if strings.HasPrefix(u, "/cosmos.safety.") {
			safetyMsgs = append(safetyMsgs, u)
		}
	}
	s.Require().ElementsMatch([]string{uSafetyParams, uSafetyTrip, uSafetyReset}, safetyMsgs, "no msg grants or delegates permissions")

	// the operator cannot appoint anyone, including itself as authority
	_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.operator,
		Params:    types.Params{Authority: s.operator, Operator: s.other, Scope: []string{types.Wildcard}},
	})
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)

	// the authority cannot trip
	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.authority, uSend))
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)

	// replacing the operator revokes the old one immediately
	_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params:    types.Params{Authority: s.authority, Operator: s.other, Scope: []string{types.Wildcard}},
	})
	s.Require().NoError(err)
	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uSend))
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.other, uSend))
	s.Require().NoError(err)
	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, uSend))
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)

	// clearing the operator leaves no address able to trip
	_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params:    types.Params{Authority: s.authority, Scope: []string{types.Wildcard}},
	})
	s.Require().NoError(err)
	for _, who := range []string{s.operator, s.other, s.authority, s.stranger} {
		_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(who, uDelegate))
		s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized, who)
	}
}

// The x/circuit authority was fixed when the app was built and checked
// with bytes.Equal. The x/safety authority lives in params and goes through
// sdk.ValidateAuthority, so it can change at runtime.
func (s *KeeperTestSuite) TestAuthorityInParams() {
	newAuth := sdk.AccAddress("unit_new_authority__").String()
	s.state(s.operator, []string{types.Wildcard}, uDelegate)

	_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.Params{Authority: newAuth}})
	s.Require().NoError(err)
	s.Require().Equal(s.authority, s.keeper.DefaultAuthority(), "the build-time default does not change")

	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.authority, uDelegate))
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized, "the build-time authority no longer resets")
	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(newAuth, uDelegate))
	s.Require().NoError(err)
}

// x/circuit revoke left state that failed validate-genesis. Every
// state reachable through the msg server must export and re-import.
func (s *KeeperTestSuite) TestGenesisRoundTrips() {
	steps := []func(){
		func() {},
		func() {
			_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uSend, "/cosmos.staking.*"))
			s.Require().NoError(err)
			_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, types.Wildcard))
			s.Require().NoError(err)
		},
		func() {
			_, err := s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, uSend))
			s.Require().NoError(err)
		},
		func() {
			// revoke the operator and narrow the scope with entries still disabled
			_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.Params{Authority: s.authority}})
			s.Require().NoError(err)
		},
		func() {
			_, err := s.msgServer.Reset(s.ctx, types.NewMsgReset(s.authority, "/cosmos.staking.*", types.Wildcard))
			s.Require().NoError(err)
		},
	}
	for i, step := range steps {
		step()
		exported, err := s.keeper.ExportGenesis(s.ctx)
		s.Require().NoError(err)
		s.Require().NoError(exported.Validate(s.keeper.AddressCodec()), "step %d", i)

		bz, err := json.Marshal(exported)
		s.Require().NoError(err)
		var decoded types.GenesisState
		s.Require().NoError(json.Unmarshal(bz, &decoded))

		savedList := s.list()
		savedParams := s.params()
		// re-import into a fresh store and continue from there
		s.SetupTest()
		s.Require().NoError(s.keeper.Params.Remove(s.ctx))
		s.Require().NoError(s.keeper.InitGenesis(s.ctx, &decoded), "step %d", i)
		again, err := s.keeper.ExportGenesis(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(savedParams, again.Params, "step %d", i)
		s.Require().ElementsMatch(savedList, again.DisabledPatterns, "step %d", i)
	}
}

// x/circuit emitted untyped events with comma-joined attributes.
// x/safety emits typed proto events whose list attributes are JSON arrays.
func (s *KeeperTestSuite) TestTypedEvents() {
	ctx := s.freshEvents()
	_, err := s.msgServer.Trip(ctx, types.NewMsgTrip(s.operator, uSend, "/cosmos.staking.*"))
	s.Require().NoError(err)

	evs := ctx.EventManager().Events()
	s.Require().Len(evs, 1)
	s.Require().Equal(proto.MessageName(&types.EventTrip{}), evs[0].Type)
	attrs := map[string]string{}
	for _, a := range evs[0].Attributes {
		attrs[a.Key] = a.Value
	}
	var patterns []string
	s.Require().NoError(json.Unmarshal([]byte(attrs["patterns"]), &patterns), "patterns must be a JSON array, got %q", attrs["patterns"])
	s.Require().Equal([]string{uSend, "/cosmos.staking.*"}, patterns)
	var operator string
	s.Require().NoError(json.Unmarshal([]byte(attrs["operator"]), &operator))
	s.Require().Equal(s.operator, operator)

	ctx = s.freshEvents()
	_, err = s.msgServer.Reset(ctx, types.NewMsgReset(s.authority, uSend, "/cosmos.staking.*"))
	s.Require().NoError(err)
	s.Require().Equal([]proto.Message{&types.EventReset{Sender: s.authority, Patterns: []string{uSend, "/cosmos.staking.*"}}}, s.typedEvents(ctx))

	ctx = s.freshEvents()
	p := types.Params{Authority: s.authority, Operator: s.other, Scope: []string{uSend}}
	_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{Authority: s.authority, Params: p})
	s.Require().NoError(err)
	s.Require().Equal([]proto.Message{&types.EventParamsUpdated{Authority: s.authority, Params: p}}, s.typedEvents(ctx))
}

// x/circuit returned unregistered errors with no ABCI code. Every x/safety
// rejection maps to a registered code: an x/safety error, or the shared SDK
// error for an unauthorized sender or an invalid address.
func (s *KeeperTestSuite) TestRejectionsUseRegisteredErrors() {
	s.state(s.operator, []string{"/cosmos.bank.*"}, uSend)
	rejections := []error{}
	add := func(_ any, err error) { rejections = append(rejections, err) }
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator)))
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, "x")))
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uMultiSend, uMultiSend)))
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip("bad", uMultiSend)))
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.stranger, uMultiSend)))
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uDelegate)))
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, "/cosmos.bank.v9.*")))
	add(s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, uSend)))
	add(s.msgServer.Reset(s.ctx, types.NewMsgReset(s.authority, uMultiSend)))
	add(s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.authority, Params: types.Params{}}))

	for _, err := range rejections {
		s.Require().Error(err)
		s.requireRegisteredError(err)
	}

	// an unauthorized sender gets the same error whichever msg it sends
	_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{Authority: s.stranger, Params: s.params()})
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
	_, err = s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.stranger, uMultiSend))
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
	_, err = s.msgServer.Reset(s.ctx, types.NewMsgReset(s.stranger, uSend))
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)

	// the breaker hands the router a registered error for a blocked msg, and
	// the router passes it on unchanged
	_, err = s.keeper.IsAllowed(s.ctx, uSend)
	s.Require().ErrorIs(err, types.ErrMsgDisabled)
	s.requireRegisteredError(err)
}

// The x/circuit disabled list query returned everything in one page.
func (s *KeeperTestSuite) TestPaginatedDisabledList() {
	var all []string
	for _, u := range s.registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		if !types.IsProtected(u) {
			all = append(all, u)
		}
	}
	s.Require().Greater(len(all), 6)
	s.state(s.operator, []string{types.Wildcard}, all...)

	var got []string
	var key []byte
	for pages := 0; ; pages++ {
		s.Require().Less(pages, len(all), "pagination does not terminate")
		res, err := s.queryClient.DisabledList(s.ctx, &types.QueryDisabledListRequest{
			Pagination: &querytypes.PageRequest{Key: key, Limit: 3, CountTotal: key == nil},
		})
		s.Require().NoError(err)
		s.Require().LessOrEqual(len(res.DisabledPatterns), 3)
		if key == nil {
			s.Require().Equal(uint64(len(all)), res.Pagination.Total)
		}
		got = append(got, res.DisabledPatterns...)
		key = res.Pagination.NextKey
		if key == nil {
			break
		}
	}
	s.Require().ElementsMatch(all, got)
}
