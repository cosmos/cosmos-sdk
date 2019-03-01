package gov

import (
	"bytes"
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// initialize the mock application for this module
func getMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []auth.Account) (
	mapp *mock.App, keeper Keeper, sk staking.Keeper, addrs []sdk.AccAddress,
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
	keeper = NewKeeper(mapp.Cdc, keyGov, pk, pk.Subspace("testgov"), ck, sk, DefaultCodespace)

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

	return mapp, keeper, sk, addrs, pubKeys, privKeys
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

// checks if two proposals are equal
func ProposalProcessEqual(proposalA ProposalProcess, proposalB ProposalProcess) bool {
	if proposalA.ProposalID == proposalB.ProposalID &&
		proposalA.GetTitle() == proposalB.GetTitle() &&
		proposalA.GetDescription() == proposalB.GetDescription() &&
		proposalA.ProposalType() == proposalB.ProposalType() &&
		proposalA.Status == proposalB.Status &&
		proposalA.FinalTallyResult.Equals(proposalB.FinalTallyResult) &&
		proposalA.SubmitTime.Equal(proposalB.SubmitTime) &&
		proposalA.DepositEndTime.Equal(proposalB.DepositEndTime) &&
		proposalA.TotalDeposit.IsEqual(proposalB.TotalDeposit) &&
		proposalA.VotingStartTime.Equal(proposalB.VotingStartTime) &&
		proposalA.VotingEndTime.Equal(proposalB.VotingEndTime) {
		return true
	}
	return false
}
