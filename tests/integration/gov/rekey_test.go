package gov_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"gotest.tools/v3/assert"
	cmp "gotest.tools/v3/assert/cmp"

	"cosmossdk.io/collections"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// govChangePubKeyMsg builds a MsgChangePubKey for acc with a valid proof of
// possession by a fresh secp256k1 key, so the only reason the msg can fail is
// the account check in the auth msg server.
func govChangePubKeyMsg(t *testing.T, ctx sdk.Context, s suite, acc sdk.AccountI) *authtypes.MsgChangePubKey {
	t.Helper()
	sk := secp256k1.GenPrivKey()
	newPk := sk.PubKey()
	anyPk, err := codectypes.NewAnyWithValue(newPk)
	assert.NilError(t, err)
	addr, err := s.AccountKeeper.AddressCodec().BytesToString(acc.GetAddress())
	assert.NilError(t, err)

	doc := authtypes.ChangePubKeyProofDoc{
		ChainId:       ctx.ChainID(),
		AccountNumber: acc.GetAccountNumber(),
		Address:       addr,
		NewPubKey:     anyPk,
	}
	signBytes, err := authtypes.ChangePubKeyProofSignBytes(s.AccountKeeper.AddressCodec(), doc, newPk)
	assert.NilError(t, err)
	sig, err := sk.Sign(signBytes)
	assert.NilError(t, err)
	proofBz, err := signing.SignatureDataToProto(&signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: sig,
	}).Marshal()
	assert.NilError(t, err)

	return &authtypes.MsgChangePubKey{Address: addr, NewPubKey: anyPk, Proof: proofBz}
}

func setupRekeyGovSuite(t *testing.T) (suite, sdk.Context, []sdk.AccAddress) {
	t.Helper()
	s := suite{}
	var err error
	s.app, err = simtestutil.SetupWithConfiguration(
		depinject.Configs(
			appConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		simtestutil.DefaultStartUpConfig(),
		&s.AccountKeeper, &s.BankKeeper, &s.DistrKeeper, &s.GovKeeper, &s.StakingKeeper, &s.cdc, &s.appBuilder,
	)
	assert.NilError(t, err)

	ctx := s.app.NewContext(false)
	addrs := simtestutil.AddTestAddrs(s.BankKeeper, s.StakingKeeper, ctx, 2, valTokens)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.app.LastBlockHeight() + 1,
		Hash:   s.app.LastCommitID().Hash,
	})
	assert.NilError(t, err)

	// Enable rekeying so a rejection can only come from the account check.
	params, err := s.AccountKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	params.PubKeyChangeEnabled = true
	assert.NilError(t, s.AccountKeeper.Params.Set(ctx, params))

	// Give addrs[0] enough bonded power to pass a proposal on its own.
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(s.StakingKeeper)
	valMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addrs[0]).String(), ed25519.GenPrivKey().PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)),
		TestDescription, TestCommissionRates, math.OneInt(),
	)
	assert.NilError(t, err)
	_, err = stakingMsgSvr.CreateValidator(ctx, valMsg)
	assert.NilError(t, err)
	_, err = s.StakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)

	return s, ctx, addrs
}

// TestGovProposalCannotRekeyGovModuleAccount checks that a passed proposal
// whose message is MsgChangePubKey for the gov module account fails on
// execution and leaves the account, PubKeyHistory and RekeyIndex unchanged.
func TestGovProposalCannotRekeyGovModuleAccount(t *testing.T) {
	s, ctx, addrs := setupRekeyGovSuite(t)
	govAddr := authtypes.NewModuleAddress(types.ModuleName)
	govAcc := s.AccountKeeper.GetAccount(ctx, govAddr)
	assert.Assert(t, govAcc != nil)
	_, isModAcc := govAcc.(sdk.ModuleAccountI)
	assert.Assert(t, isModAcc)
	assert.Assert(t, govAcc.GetPubKey() == nil)

	msg := govChangePubKeyMsg(t, ctx, s, govAcc)
	proposal, err := s.GovKeeper.SubmitProposal(ctx, []sdk.Msg{msg}, "", "rekey gov", "rekey the gov module account", addrs[0], false)
	assert.NilError(t, err)

	govMsgSvr := keeper.NewMsgServerImpl(s.GovKeeper)
	deposit := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, s.StakingKeeper.TokensFromConsensusPower(ctx, 10)))
	_, err = govMsgSvr.Deposit(ctx, v1.NewMsgDeposit(addrs[0], proposal.Id, deposit))
	assert.NilError(t, err)
	assert.NilError(t, s.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

	params, err := s.GovKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	header := ctx.BlockHeader()
	header.Time = header.Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
	ctx = ctx.WithBlockHeader(header)

	assert.NilError(t, gov.EndBlocker(ctx, s.GovKeeper))

	attrs, ok := ctx.EventManager().Events().GetAttributes(types.AttributeKeyProposalLog)
	assert.Assert(t, ok)
	assert.Assert(t, cmp.Contains(attrs[len(attrs)-1].Value, "failed on execution"))

	proposal, err = s.GovKeeper.Proposals.Get(ctx, proposal.Id)
	assert.NilError(t, err)
	assert.Equal(t, v1.StatusFailed, proposal.Status)
	assert.Assert(t, cmp.Contains(proposal.FailedReason, authtypes.ErrAccountNotRekeyable.Error()))
	// The gov module account has no pubkey, so the "has no pubkey" check would
	// also reject this msg. This string is what pins the ModuleAccountI check;
	// TestChangePubKey_ModuleAccountWithPubKey in x/auth/keeper covers a module
	// account that does have a pubkey.
	assert.Assert(t, cmp.Contains(proposal.FailedReason, "is a module account"))

	after := s.AccountKeeper.GetAccount(ctx, govAddr)
	_, isModAcc = after.(sdk.ModuleAccountI)
	assert.Assert(t, isModAcc)
	assert.Assert(t, after.GetPubKey() == nil)
	assertNoRekeyState(t, ctx, s, govAddr)
}

// TestGovProposalCannotRekeyUserAccount checks that a proposal cannot carry a
// MsgChangePubKey for a user account: its signer is not the gov module
// account, so submission fails and nothing is stored.
func TestGovProposalCannotRekeyUserAccount(t *testing.T) {
	s, ctx, addrs := setupRekeyGovSuite(t)
	userAcc := s.AccountKeeper.GetAccount(ctx, addrs[1])
	assert.Assert(t, userAcc != nil)
	// Give the account a real pubkey so x/auth's "has no pubkey" check cannot
	// mask a regression in gov's signer check.
	oldPk := secp256k1.GenPrivKey().PubKey()
	assert.NilError(t, userAcc.SetPubKey(oldPk))
	s.AccountKeeper.SetAccount(ctx, userAcc)

	msg := govChangePubKeyMsg(t, ctx, s, userAcc)
	_, err := s.GovKeeper.SubmitProposal(ctx, []sdk.Msg{msg}, "", "rekey user", "rekey a user account", addrs[0], false)
	assert.ErrorIs(t, err, types.ErrInvalidSigner)

	after := s.AccountKeeper.GetAccount(ctx, addrs[1])
	assert.Assert(t, after.GetPubKey() != nil)
	assert.Assert(t, oldPk.Equals(after.GetPubKey()))
	assertNoRekeyState(t, ctx, s, addrs[1])
}

func assertNoRekeyState(t *testing.T, ctx sdk.Context, s suite, addr sdk.AccAddress) {
	t.Helper()
	iter, err := s.AccountKeeper.PubKeyHistory.Iterate(ctx, collections.NewPrefixedPairRange[sdk.AccAddress, uint64](addr))
	assert.NilError(t, err)
	defer iter.Close()
	assert.Assert(t, !iter.Valid(), "unexpected PubKeyHistory entry for %s", addr)

	idx, err := s.AccountKeeper.RekeyIndex.Iterate(ctx, nil)
	assert.NilError(t, err)
	defer idx.Close()
	assert.Assert(t, !idx.Valid(), "unexpected RekeyIndex entry")
}
