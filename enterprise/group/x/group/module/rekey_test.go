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

package module_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	group "github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group/keeper"
	grouptestutil "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// TestGroupProposalCannotRekeyPolicyAccount checks that a passed group
// proposal whose message is MsgChangePubKey for its own group policy account
// fails on execution and leaves the policy account's ModuleCredential pubkey,
// PubKeyHistory and RekeyIndex unchanged.
func TestGroupProposalCannotRekeyPolicyAccount(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
		groupKeeper   keeper.Keeper
	)
	app, err := simtestutil.Setup(
		depinject.Configs(
			grouptestutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&accountKeeper, &bankKeeper, &stakingKeeper, &groupKeeper,
	)
	require.NoError(t, err)

	ctx := app.NewContext(false).WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, math.NewInt(30000000))
	addrCodec := accountKeeper.AddressCodec()
	member, err := addrCodec.BytesToString(addrs[0])
	require.NoError(t, err)

	// Enable rekeying so a rejection can only come from the account check.
	params, err := accountKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.PubKeyChangeEnabled = true
	require.NoError(t, accountKeeper.Params.Set(ctx, params))

	groupRes, err := groupKeeper.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   member,
		Members: []group.MemberRequest{{Address: member, Weight: "1"}},
	})
	require.NoError(t, err)
	policyReq := &group.MsgCreateGroupPolicy{Admin: member, GroupId: groupRes.GroupId}
	require.NoError(t, policyReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Second, 0)))
	policyRes, err := groupKeeper.CreateGroupPolicy(ctx, policyReq)
	require.NoError(t, err)
	policyAddr, err := addrCodec.StringToBytes(policyRes.Address)
	require.NoError(t, err)

	policyAcc := accountKeeper.GetAccount(ctx, policyAddr)
	require.NotNil(t, policyAcc)
	oldPk := policyAcc.GetPubKey()
	require.IsType(t, &authtypes.ModuleCredential{}, oldPk)

	// Build the msg with a valid proof of possession by a fresh key, so the
	// only reason it can fail is the account check in the auth msg server.
	sk := secp256k1.GenPrivKey()
	anyPk, err := codectypes.NewAnyWithValue(sk.PubKey())
	require.NoError(t, err)
	doc := authtypes.ChangePubKeyProofDoc{
		ChainId:       ctx.ChainID(),
		AccountNumber: policyAcc.GetAccountNumber(),
		Address:       policyRes.Address,
		NewPubKey:     anyPk,
	}
	signBytes, err := authtypes.ChangePubKeyProofSignBytes(addrCodec, doc, sk.PubKey())
	require.NoError(t, err)
	sig, err := sk.Sign(signBytes)
	require.NoError(t, err)
	proofBz, err := signing.SignatureDataToProto(&signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: sig,
	}).Marshal()
	require.NoError(t, err)
	msg := &authtypes.MsgChangePubKey{Address: policyRes.Address, NewPubKey: anyPk, Proof: proofBz}

	propReq := &group.MsgSubmitProposal{GroupPolicyAddress: policyRes.Address, Proposers: []string{member}}
	require.NoError(t, propReq.SetMsgs([]sdk.Msg{msg}))
	propRes, err := groupKeeper.SubmitProposal(ctx, propReq)
	require.NoError(t, err)
	_, err = groupKeeper.Vote(ctx, &group.MsgVote{ProposalId: propRes.ProposalId, Voter: member, Option: group.VOTE_OPTION_YES})
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second)).WithEventManager(sdk.NewEventManager())
	execRes, err := groupKeeper.Exec(ctx, &group.MsgExec{Executor: member, ProposalId: propRes.ProposalId})
	require.NoError(t, err)
	require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_FAILURE, execRes.Result)

	var execEvent *group.EventExec
	for _, ev := range ctx.EventManager().Events() {
		if ev.Type != proto.MessageName(&group.EventExec{}) {
			continue
		}
		parsed, err := sdk.ParseTypedEvent(abci.Event(ev))
		require.NoError(t, err)
		execEvent = parsed.(*group.EventExec)
	}
	require.NotNil(t, execEvent)
	require.Contains(t, execEvent.Logs, authtypes.ErrAccountNotRekeyable.Error())
	require.Contains(t, execEvent.Logs, "is a module credential account")

	after := accountKeeper.GetAccount(ctx, policyAddr)
	require.True(t, oldPk.Equals(after.GetPubKey()), "policy account pubkey changed")

	hasHistory := false
	err = accountKeeper.PubKeyHistory.Walk(ctx, collections.NewPrefixedPairRange[sdk.AccAddress, uint64](policyAddr),
		func(collections.Pair[sdk.AccAddress, uint64], authtypes.PubKeyHistoryEntry) (bool, error) {
			hasHistory = true
			return true, nil
		})
	require.NoError(t, err)
	require.False(t, hasHistory, "unexpected PubKeyHistory entry")
	has, err := accountKeeper.RekeyIndex.Has(ctx, collections.Join(sdk.AccAddress(sk.PubKey().Address()), sdk.AccAddress(policyAddr)))
	require.NoError(t, err)
	require.False(t, has, "unexpected RekeyIndex entry")
}
