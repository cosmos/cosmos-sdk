package gov

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

type testInput struct {
	mApp     *mock.App
	keeper   Keeper
	router   Router
	sk       staking.Keeper
	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

func getMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []auth.Account) testInput {
	mApp := mock.NewApp()

	staking.RegisterCodec(mApp.Cdc)
	RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tKeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyGov := sdk.NewKVStoreKey(StoreKey)

	pk := mApp.ParamsKeeper

	rtr := NewRouter().
		AddRoute(RouterKey, ProposalHandler)

	ck := bank.NewBaseKeeper(mApp.AccountKeeper, mApp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := staking.NewKeeper(mApp.Cdc, keyStaking, tKeyStaking, ck, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	keeper := NewKeeper(mApp.Cdc, keyGov, pk, pk.Subspace("testgov"), ck, sk, DefaultCodespace, rtr)

	mApp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mApp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mApp.SetEndBlocker(getEndBlocker(keeper))
	mApp.SetInitChainer(getInitChainer(mApp, keeper, sk, mApp.AccountKeeper, genState))

	require.NoError(t, mApp.CompleteSetup(keyStaking, tKeyStaking, keyGov))

	valTokens := sdk.TokensFromTendermintPower(42)

	var (
		addrs    []sdk.AccAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
	}

	mock.SetGenesis(mApp, genAccs)

	return testInput{mApp, keeper, rtr, sk, addrs, pubKeys, privKeys}
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
func getInitChainer(mapp *mock.App, keeper Keeper, stakingKeeper staking.Keeper,
	accountKeeper staking.AccountKeeper, genState GenesisState) sdk.InitChainer {

	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakingGenesis := staking.DefaultGenesisState()
		tokens := sdk.TokensFromTendermintPower(100000)
		stakingGenesis.Pool.NotBondedTokens = tokens

		validators := staking.InitGenesis(ctx, stakingKeeper, accountKeeper, stakingGenesis)
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

func testProposal() Content {
	return NewTextProposal("Test", "description")
}

const contextKeyBadProposal = "contextKeyBadProposal"

// badProposalHandler implements a governance proposal handler that is identical
// to the actual handler except this fails if the context doesn't contain a value
// for the key contextKeyBadProposal or if the value is false.
func badProposalHandler(ctx sdk.Context, c Content) sdk.Error {
	switch c.ProposalType() {
	case ProposalTypeText, ProposalTypeSoftwareUpgrade:
		v := ctx.Value(contextKeyBadProposal)

		if v == nil || !v.(bool) {
			return sdk.ErrInternal("proposal failed")
		}

		return nil

	default:
		errMsg := fmt.Sprintf("unrecognized gov proposal type: %s", c.ProposalType())
		return sdk.ErrUnknownRequest(errMsg)
	}
}

// checks if two proposals are equal (note: slow, for tests only)
func ProposalEqual(proposalA Proposal, proposalB Proposal) bool {
	return bytes.Equal(types.ModuleCdc.MustMarshalBinaryBare(proposalA),
		types.ModuleCdc.MustMarshalBinaryBare(proposalB))
}

var (
	pubkeys = []crypto.PubKey{
		ed25519.GenPrivKey().PubKey(),
		ed25519.GenPrivKey().PubKey(),
		ed25519.GenPrivKey().PubKey(),
	}

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
