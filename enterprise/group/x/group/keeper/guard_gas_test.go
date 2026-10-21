// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper_test

import (
	"context"
	"fmt"
	"maps"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/codec/address"
	group "github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group/keeper"
	module2 "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/module"
	grouptestutil "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/testutil"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// guardLookupGas is the gas meteredBreaker charges for one lookup.
const guardLookupGas = 1_000

// meteredBreaker allows every msg except blocked, charges guardLookupGas to
// the Context gas meter for every lookup, as a store-backed breaker such as
// x/safety pays for its reads, and counts lookups per type URL.
type meteredBreaker struct {
	calls   map[string]int
	blocked string
}

func (b *meteredBreaker) IsAllowed(ctx context.Context, typeURL string) (bool, error) {
	sdk.UnwrapSDKContext(ctx).GasMeter().ConsumeGas(guardLookupGas, "circuit breaker lookup")
	b.calls[typeURL]++
	return typeURL != b.blocked, nil
}

// guardedCounterServer handles MsgCounter after calling
// msgservice.CheckMsgAllowed, as SDK msg servers do, and counts its runs.
type guardedCounterServer struct {
	runs *int
}

func (s guardedCounterServer) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter) (*baseapptestutil.MsgCreateCounterResponse, error) {
	if err := msgservice.CheckMsgAllowed(ctx, msg); err != nil {
		return nil, err
	}
	*s.runs++
	return &baseapptestutil.MsgCreateCounterResponse{}, nil
}

// newGuardedSuite returns a group TestSuite whose keeper is registered on a
// MsgServiceRouter that checks every msg against a meteredBreaker.
func newGuardedSuite(t *testing.T) (*TestSuite, *baseapp.MsgServiceRouter, *meteredBreaker, moduletestutil.TestEncodingConfig) {
	t.Helper()
	s := &TestSuite{}
	s.SetT(t)

	key := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module2.AppModuleBasic{})
	baseapptestutil.RegisterInterfaces(encCfg.InterfaceRegistry)
	s.addrs = simtestutil.CreateIncrementalAccounts(1)

	s.accountKeeper = grouptestutil.NewMockAccountKeeper(gomock.NewController(t))
	s.accountKeeper.EXPECT().GetAccount(gomock.Any(), s.addrs[0]).Return(authtypes.NewBaseAccountWithAddress(s.addrs[0])).AnyTimes()
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	cb := &meteredBreaker{calls: map[string]int{}}
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	router.SetCircuit(cb)
	s.groupKeeper = keeper.NewKeeper(key, encCfg.Codec, router, s.accountKeeper, group.DefaultConfig())
	group.RegisterMsgServer(router, s.groupKeeper)

	s.blockTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.sdkCtx = testCtx.Ctx.WithBlockTime(s.blockTime)
	s.ctx = s.sdkCtx
	return s, router, cb, encCfg
}

// TestExecInnerMsgGasMatchesBreakerFreeGuard shows that the msg server guard
// adds no gas to the msgs a group proposal executes, which group dispatches
// through the MsgServiceRouter. The baseline runs the same group MsgExec with
// the breaker removed from the Context, so the router checks each msg as
// before and every guard is a no-op.
func TestExecInnerMsgGasMatchesBreakerFreeGuard(t *testing.T) {
	s, router, cb, encCfg := newGuardedSuite(t)
	runs := 0
	baseapptestutil.RegisterCounterServer(router, guardedCounterServer{runs: &runs})

	policyAddr, _ := s.createGroupAndGroupPolicy(s.addrs[0], nil, group.NewThresholdDecisionPolicy("1", time.Second, 0))
	msgs := []sdk.Msg{
		&baseapptestutil.MsgCounter{Counter: 1, Signer: policyAddr},
		&baseapptestutil.MsgCounter{Counter: 2, Signer: policyAddr},
	}
	id := submitProposalAndVoteWithPolicy(s.ctx, s, policyAddr, msgs, []string{s.addrs[0].String()}, group.VOTE_OPTION_YES)

	exec := &group.MsgExec{ProposalId: id, Executor: s.addrs[0].String()}
	execURL := sdk.MsgTypeURL(exec)
	counterURL := sdk.MsgTypeURL(msgs[0])

	run := func(guardOff bool) (uint64, map[string]int) {
		ctx, _ := s.sdkCtx.WithBlockTime(s.blockTime.Add(2 * time.Second)).CacheContext()
		ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).WithCircuitBreaker(cb)
		if guardOff {
			ctx = ctx.WithCircuitBreaker(nil)
		}
		cb.calls = map[string]int{}
		runs = 0

		res, err := router.Handler(exec)(ctx, exec)
		s.Require().NoError(err)
		s.Require().Len(res.MsgResponses, 1)
		var resp group.MsgExecResponse
		s.Require().NoError(encCfg.Codec.Unmarshal(res.MsgResponses[0].Value, &resp))
		s.Require().Equal(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS, resp.Result)
		s.Require().Equal(2, runs)
		return ctx.GasMeter().GasConsumed(), maps.Clone(cb.calls)
	}

	baselineGas, baselineCalls := run(true)
	guardedGas, guardedCalls := run(false)
	s.Require().Equal(map[string]int{execURL: 1, counterURL: 2}, guardedCalls,
		"the router checks MsgExec and each proposal msg once, and the guard repeats none")
	s.Require().Equal(baselineCalls, guardedCalls)
	s.Require().Equal(baselineGas, guardedGas)
}

// TestCreateGroupWithPolicyChecksItsInnerCalls shows that a routed
// MsgCreateGroupWithPolicy pays a guard lookup for each msg server it calls
// directly, and that an exact trip of any of those msgs fails it.
func TestCreateGroupWithPolicyChecksItsInnerCalls(t *testing.T) {
	for _, policyAsAdmin := range []bool{true, false} {
		t.Run(fmt.Sprintf("group_policy_as_admin=%t", policyAsAdmin), func(t *testing.T) {
			s, router, cb, _ := newGuardedSuite(t)
			s.setNextAccount()

			admin := s.addrs[0].String()
			msg, err := group.NewMsgCreateGroupWithPolicy(admin, []group.MemberRequest{{Address: admin, Weight: "1"}},
				"", "", policyAsAdmin, group.NewThresholdDecisionPolicy("1", time.Second, 0))
			s.Require().NoError(err)

			// The admin updates run only when the policy becomes the admin.
			called := []string{
				sdk.MsgTypeURL(&group.MsgCreateGroup{}),
				sdk.MsgTypeURL(&group.MsgCreateGroupPolicy{}),
			}
			adminUpdates := []string{
				sdk.MsgTypeURL(&group.MsgUpdateGroupAdmin{}),
				sdk.MsgTypeURL(&group.MsgUpdateGroupPolicyAdmin{}),
			}
			notCalled := adminUpdates
			if policyAsAdmin {
				called = append(called, adminUpdates...)
				notCalled = nil
			}

			run := func(blocked string) error {
				ctx, _ := s.sdkCtx.CacheContext()
				ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()).WithCircuitBreaker(cb)
				cb.calls = map[string]int{}
				cb.blocked = blocked
				_, err := router.Handler(msg)(ctx, msg)
				return err
			}

			s.Require().NoError(run(""))
			want := map[string]int{sdk.MsgTypeURL(msg): 1}
			for _, url := range called {
				want[url] = 1
			}
			s.Require().Equal(want, cb.calls)

			for _, url := range called {
				err := run(url)
				s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized, url)
				s.Require().ErrorContains(err, url)
			}
			for _, url := range notCalled {
				s.Require().NoError(run(url), url)
			}
		})
	}
}
