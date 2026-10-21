package keeper_test

import (
	"encoding/json"
	"strings"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/safety"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

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
	s.requireRegisteredError(types.ErrMaxDepth.Wrap("x"))
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

// The x/circuit CLI examples could not be run as written. Every
// x/safety trip and reset example must parse into patterns this keeper
// accepts, one pattern per argument.
func (s *KeeperTestSuite) TestCLIExamplesRun() {
	examples := map[string]string{}
	for _, rpc := range safety.NewAppModule(s.keeper).AutoCLIOptions().Tx.RpcCommandOptions {
		examples[rpc.RpcMethod] = rpc.Example
	}
	args := func(example, verb string) [][]string {
		var out [][]string
		for line := range strings.SplitSeq(example, "\n") {
			fields := strings.Fields(line)
			var patterns []string
			on := false
			for _, f := range fields {
				if strings.HasPrefix(f, "--") {
					break
				}
				if on {
					patterns = append(patterns, strings.Trim(f, `'`))
				}
				on = on || f == verb
			}
			s.Require().NotEmpty(patterns, "example line %q has no patterns", line)
			out = append(out, patterns)
		}
		return out
	}

	trips := args(examples["Trip"], "trip")
	resets := args(examples["Reset"], "reset")
	s.Require().NotEmpty(trips)
	s.Require().NotEmpty(resets)

	for _, patterns := range trips {
		_, err := s.msgServer.Trip(s.ctx, types.NewMsgTrip(s.operator, patterns...))
		s.Require().NoError(err, "trip example %v", patterns)
	}
	for _, patterns := range resets {
		_, err := s.msgServer.Reset(s.ctx, types.NewMsgReset(s.operator, patterns...))
		s.Require().NoError(err, "reset example %v", patterns)
	}
	s.Require().Empty(s.list())

	// the is-allowed example must name a type URL the query accepts
	var isAllowed string
	for _, rpc := range safety.NewAppModule(s.keeper).AutoCLIOptions().Query.RpcCommandOptions {
		if rpc.RpcMethod == "IsAllowed" {
			isAllowed = rpc.Example
		}
	}
	fields := strings.Fields(isAllowed)
	s.Require().NotEmpty(fields, "is-allowed example")
	res, err := s.queryClient.IsAllowed(s.ctx, &types.QueryIsAllowedRequest{MsgTypeUrl: fields[len(fields)-1]})
	s.Require().NoError(err, "is-allowed example %q", isAllowed)
	s.Require().True(res.Allowed)
}
