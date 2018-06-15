package simpleGovernance

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/simpleGov/app"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// dummy addresses used for testing
var (
	addrs = []sdk.Address{
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
	pks = []crypto.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB56"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB58"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB59"),
	}

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

	coinsHandlerTest = []sdk.Coins{
		sdk.Coins{{"Atom", 101}, {"eth", 20}}, // ok
		sdk.Coins{{"eth", 10}, {"Atom", 0}},   // empty coins
		sdk.Coins{{"BTC", 15}, {"Atom", 50}},  // balance below deposit
	}

	options = []string{
		"Yes",
		"No",
		"Abstain",
		"",
		"          ",
	}
	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey
)

//_______________________________________________________________________________________

// // intended to be used with require/assert:  require.True(ValEq(...))
// func valEq(t *testing.T, exp, got Validator) (*testing.T, bool, string, Validator, Validator) {
// 	return t, exp.equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
// }

func loggerAndDB() (log.Logger, db.DB) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	dB := db.NewMemDB()
	return logger, dB
}

func newSimpleGovApp() *app.SimpleGovApp {
	logger, dB := loggerAndDB()
	return app.NewSimpleGovApp(logger, dB)
}

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.Address) {
	key := crypto.GenPrivKeyEd25519()
	pub := key.PubKey()
	addr := sdk.Address(pub.Address())
	return key, pub, addr
}

func initialPool() stake.Pool {
	return stake.Pool{
		LooseUnbondedTokens:     0,
		BondedTokens:            0,
		UnbondingTokens:         0,
		UnbondedTokens:          0,
		BondedShares:            sdk.ZeroRat(),
		UnbondingShares:         sdk.ZeroRat(),
		UnbondedShares:          sdk.ZeroRat(),
		InflationLastTime:       0,
		Inflation:               sdk.NewRat(7, 100),
		DateLastCommissionReset: 0,
		PrevBondedShares:        sdk.ZeroRat(),
	}
}

func defaultParams() stake.Params {
	return stake.Params{
		InflationRateChange: sdk.NewRat(13, 100),
		InflationMax:        sdk.NewRat(20, 100),
		InflationMin:        sdk.NewRat(7, 100),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}

//_______________________________________________________________________________________

func makeTestCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.MsgSend{}, "test/stake/Send", nil)
	cdc.RegisterConcrete(bank.MsgIssue{}, "test/stake/Issue", nil)
	cdc.RegisterConcrete(SubmitProposalMsg{}, "simple_governance/SubmitProposalMsg", nil)
	cdc.RegisterConcrete(VoteMsg{}, "simple_governance/VoteMsg", nil)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/stake/Account", nil)
	wire.RegisterCrypto(cdc)

	return cdc
}

// hogpodge of all sorts of input required for testing
func createTestInput(t *testing.T, initCoins int64) (sdk.Context, auth.AccountMapper, Keeper) {
	app := newSimpleGovApp()
	keyStake := sdk.NewKVStoreKey("stake")
	keyAuth := sdk.NewKVStoreKey("auth")
	simpleGovKey := sdk.NewKVStoreKey("simpleGov")

	// db := dbm.NewMemDB()

	db := dbm.NewDB("filesystemDB", dbm.FSDBBackend, "dir")
	ms := store.NewCommitMultiStore(db)

	app.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	app.MountStoreWithDB(keyAuth, sdk.StoreTypeIAVL, db)
	app.MountStoreWithDB(simpleGovKey, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := app.NewContext(isCheckTx, abci.Header{ChainID: "foochainid"})
	app.Mo
	accountMapper := auth.NewAccountMapper(
		cdc,                 // amino codec
		keyAuth,             // target store
		&auth.BaseAccount{}, // prototype
	)
	ck := bank.NewKeeper(accountMapper)
	stakeKeeper := stake.NewKeeper(cdc, keyStake, ck, DefaultCodespace)

	// fill all the addresses with some coins
	for _, addr := range addrs {
		ck.AddCoins(ctx, addr, sdk.Coins{
			{"Atom", initCoins},
		})
	}

	keeper := NewKeeper(simpleGovKey, ck, stakeKeeper, DefaultCodespace)
	return ctx, accountMapper, keeper
}

func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	//res, err = crypto.PubKeyFromBytes(pkBytes)
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

// for incode address generation
func testAddr(addr string) sdk.Address {
	res, err := sdk.GetAccAddressHex(addr)
	if err != nil {
		panic(err)
	}
	return res
}
