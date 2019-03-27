package gov

import (
	"bytes"
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// initialize the mock application for this module
func GetMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []auth.Account) (
	mapp *mock.App, keeper Keeper, router Router, sk staking.Keeper, addrs []sdk.AccAddress,
	pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {

	mapp = mock.NewApp()

	staking.RegisterCodec(mapp.Cdc)
	RegisterCodec(mapp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyGov := sdk.NewKVStoreKey(StoreKey)

	pk := mapp.ParamsKeeper
	ck := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk = staking.NewKeeper(mapp.Cdc, keyStaking, tkeyStaking, ck, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	keeper, router = NewKeeper(mapp.Cdc, keyGov, pk, pk.Subspace("testgov"), ck, sk, DefaultCodespace)

	router.AddRoute(RouterKey, ProposalHandler)

	mapp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mapp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mapp.SetEndBlocker(getEndBlocker(keeper))
	mapp.SetInitChainer(getInitChainer(mapp, keeper, sk, genState))

	require.NoError(t, mapp.CompleteSetup(keyStaking, tkeyStaking, keyGov))

	valTokens := sdk.TokensFromTendermintPower(42)
	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
	}

	mock.SetGenesis(mapp, genAccs)

	return mapp, keeper, router, sk, addrs, pubKeys, privKeys
}

// gov and staking endblocker
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		tags := EndBlocker(ctx, keeper)
		return abci.ResponseEndBlock{
			Tags: tags,
		}
	}
}

// gov and staking initchainer
func getInitChainer(mapp *mock.App, keeper Keeper, stakingKeeper staking.Keeper, genState GenesisState) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakingGenesis := staking.DefaultGenesisState()
		tokens := sdk.TokensFromTendermintPower(100000)
		stakingGenesis.Pool.NotBondedTokens = tokens

		validators, err := staking.InitGenesis(ctx, stakingKeeper, stakingGenesis)
		if err != nil {
			panic(err)
		}
		if genState.IsEmpty() {
			InitGenesis(ctx, keeper, DefaultGenesisState())
		} else {
			InitGenesis(ctx, keeper, genState)
		}
		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

// TODO: Remove once address interface has been implemented (ref: #2186)
func SortValAddresses(addrs []sdk.ValAddress) {
	var byteAddrs [][]byte
	for _, addr := range addrs {
		byteAddrs = append(byteAddrs, addr.Bytes())
	}

	SortByteArrays(byteAddrs)

	for i, byteAddr := range byteAddrs {
		addrs[i] = byteAddr
	}
}

// Sorts Addresses
func SortAddresses(addrs []sdk.AccAddress) {
	var byteAddrs [][]byte
	for _, addr := range addrs {
		byteAddrs = append(byteAddrs, addr.Bytes())
	}
	SortByteArrays(byteAddrs)
	for i, byteAddr := range byteAddrs {
		addrs[i] = byteAddr
	}
}

// implement `Interface` in sort package.
type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// Public
func SortByteArrays(src [][]byte) [][]byte {
	sorted := sortByteArrays(src)
	sort.Sort(sorted)
	return sorted
}

var (
	pubkeys = []crypto.PubKey{ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey()}

	testDescription   = staking.NewDescription("T", "E", "S", "T")
	testCommissionMsg = staking.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func createValidators(t *testing.T, stakingHandler sdk.Handler, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {

		valTokens := sdk.TokensFromTendermintPower(powerAmt[i])
		valCreateMsg := staking.NewMsgCreateValidator(
			addrs[i], pubkeys[i], sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			testDescription, testCommissionMsg, sdk.OneInt(),
		)

		res := stakingHandler(ctx, valCreateMsg)
		require.True(t, res.IsOK())
	}
}

func testProposal() proposal.Content {
	return NewTextProposal("Test", "description")
}

// checks if two proposals are equal (note: slow, for tests only)
func ProposalEqual(proposalA Proposal, proposalB Proposal) bool {
	return bytes.Equal(msgCdc.MustMarshalBinaryBare(proposalA), msgCdc.MustMarshalBinaryBare(proposalB))
}

// SubmitProposal which returns Proposal instead of ProposalID
func testSubmitProposal(ctx sdk.Context, k Keeper, content proposal.Content) (p Proposal, err sdk.Error) {
	var id uint64
	id, err = k.SubmitProposal(ctx, content)
	if err != nil {
		return
	}

	var ok bool
	p, ok = k.GetProposal(ctx, id)
	if !ok {
		err = sdk.NewError(sdk.CodespaceUndefined, sdk.CodeType(1), "Invalid ProposalID in testSubmitProposal; id=%d", id)
	}
	return
}

// XXX: check if it is secure to export
// calling handlers is not vulnearable, it is possible to do anyway
// k.ck.SetSendEnabled might be vulnearable
// using predefined address might be vulnearable(createValidators, k.AddVote)
func TestProposal(t *testing.T, mapp *mock.App, addr sdk.AccAddress, k Keeper, sk staking.Keeper, content proposal.Content) sdk.Tags {
	handler := NewHandler(k)
	stakingHandler := staking.NewHandler(sk)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	valAddr := sdk.ValAddress(addr)

	k.ck.SetSendEnabled(ctx, true)
	createValidators(t, stakingHandler, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	staking.EndBlocker(ctx, sk)

	id, err := k.SubmitProposal(ctx, content)
	require.NoError(t, err)

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(10))}
	newDepositMsg := NewMsgDeposit(addr, id, proposalCoins)
	res := handler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	err = k.AddVote(ctx, id, addr, OptionYes)
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(k.GetDepositParams(ctx).MaxDepositPeriod).Add(k.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	return EndBlocker(ctx, k)
}
