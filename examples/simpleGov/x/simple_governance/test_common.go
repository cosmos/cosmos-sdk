package simpleGovernance

import (
	"bytes"
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// dummy addresses used for testing
var (
	addrs = []sdk.AccAddress{
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6160"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6161"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6162"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6163"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6164"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6165"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6166"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6167"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6168"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6169"),
	}

	// dummy pubkeys used for testing
	// pks = []crypto.PubKey{
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB56"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB58"),
	// 	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB59"),
	// }

	titles = []string{
		"Photons at launch",
		"IBC integration at launch",
		"Voting Period update",
		"Validator Set migration",
		"Upgrade Zone to a Hub",
		"Other Fees",
		"Change downtime parameter",
		"Change validator limit",
	}

	descriptions = []string{
		"Should we include Photons at launch?",
		"Should we include IBC integration at launch?",
		"Should we update the proposal voting period to 13000000 blocks?",
		"Should we change from the default Cosmos Hub validator set to our own zone validator set?",
		"Should we upgrade our zone to convert it to an independent Hub?",
		"Should we accept coins from other zones as fees?",
		"Should we change downtime to 50% of the last 50 blocks instead of last 100 blocks?",
		"Should we change the number of validators to 500?",
	}

	// coinsHandlerTest = []sdk.Coins{
	// 	sdk.Coins{{"Atom", sdk.NewInt(int64(101))}, {"eth", sdk.NewInt(int64(20))}}, // ok
	// 	sdk.Coins{{"eth", sdk.NewInt(int64(10))}, {"Atom", sdk.NewInt(int64(0))}},   // empty coins
	// 	sdk.Coins{{"BTC", sdk.NewInt(int64(15))}, {"Atom", sdk.NewInt(int64(50))}},  // balance below deposit
	// }

	options = []string{
		"Yes",
		"No",
		"Abstain",
		"",
		"          ",
	}
	emptyAddr   sdk.AccAddress
	emptyPubkey crypto.PubKey
)

// func newPubKey(pk string) (res crypto.PubKey) {
// 	pkBytes, err := hex.DecodeString(pk)
// 	if err != nil {
// 		panic(err)
// 	}
// 	//res, err = crypto.PubKeyFromBytes(pkBytes)
// 	var pkEd crypto.PubKeyEd25519
// 	copy(pkEd[:], pkBytes[:])
// 	return pkEd
// }

// for incode address generation
func testAddr(addr string) sdk.AccAddress {
	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		panic(err)
	}
	return res
}

// initialize the mock application for this module
func getMockApp(t *testing.T, numGenAccs int) (*mock.App, Keeper, stake.Keeper, []sdk.AccAddress, []crypto.PubKey, []crypto.PrivKey) {
	mapp := mock.NewApp()

	stake.RegisterWire(mapp.Cdc)
	RegisterWire(mapp.Cdc)

	keyGlobalParams := sdk.NewKVStoreKey("params")
	keyStake := sdk.NewKVStoreKey("stake")
	keyGov := sdk.NewKVStoreKey("simplegov")

	ck := bank.NewKeeper(mapp.AccountMapper)
	sk := stake.NewKeeper(mapp.Cdc, keyStake, ck, mapp.RegisterCodespace(stake.DefaultCodespace))
	keeper := NewKeeper(mapp.Cdc, keyGov, ck, sk, mapp.RegisterCodespace(stake.DefaultCodespace))
	mapp.Router().AddRoute("simplegov", NewHandler(keeper))

	mapp.SetEndBlocker(NewEndBlocker(keeper))
	mapp.SetInitChainer(getInitChainer(mapp, keeper, sk))

	require.NoError(t, mapp.CompleteSetup([]*sdk.KVStoreKey{keyStake, keyGov, keyGlobalParams}))

	genAccs, addresses, pubKeys, privKeys := mock.CreateGenAccounts(numGenAccs, sdk.Coins{sdk.NewInt64Coin("steak", 42)})

	mock.SetGenesis(mapp, genAccs)

	return mapp, keeper, sk, addresses, pubKeys, privKeys
}

// gov and stake initchainer
func getInitChainer(mapp *mock.App, keeper Keeper, stakeKeeper stake.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakeGenesis := stake.DefaultGenesisState()
		stakeGenesis.Pool.LooseTokens = sdk.NewRat(100000)

		validators, err := stake.InitGenesis(ctx, stakeKeeper, stakeGenesis)
		if err != nil {
			panic(err)
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
