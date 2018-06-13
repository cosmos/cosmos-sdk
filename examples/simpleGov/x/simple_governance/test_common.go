package simpleGovernance

// TODO
// import (
// 	"encoding/hex"
// 	"testing"
//
// 	"github.com/stretchr/testify/require"
//
// 	abci "github.com/tendermint/abci/types"
// 	crypto "github.com/tendermint/go-crypto"
// 	dbm "github.com/tendermint/tmlibs/db"
// 	"github.com/tendermint/tmlibs/log"
//
// 	"github.com/cosmos/cosmos-sdk/store"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/wire"
// 	"github.com/cosmos/cosmos-sdk/x/auth"
// 	"github.com/cosmos/cosmos-sdk/x/bank"
// 	"github.com/cosmos/cosmos-sdk/x/stake"
// )
//
// // dummy addresses used for testing
// var (
// 	addrs = []sdk.Address{
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6160"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6161"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6162"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6163"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6164"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6165"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6166"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6167"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6168"),
// 		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6169"),
// 	}
//
// 	// dummy pubkeys used for testing
// 	pks = []crypto.PubKey{
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB56"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB58"),
// 		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB59"),
// 	}
//
// 	emptyAddr   sdk.Address
// 	emptyPubkey crypto.PubKey
// )
//
// //_______________________________________________________________________________________
//
// // intended to be used with require/assert:  require.True(ValEq(...))
// func ValEq(t *testing.T, exp, got Validator) (*testing.T, bool, string, Validator, Validator) {
// 	return t, exp.equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
// }
//
// func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.Address) {
// 	key := crypto.GenPrivKeyEd25519()
// 	pub := key.PubKey()
// 	addr := sdk.Address(pub.Address())
// 	return key, pub, addr
// }
//
// func initialPool() stake.Pool {
// 	return stake.Pool{
// 		LooseUnbondedTokens:     0,
// 		BondedTokens:            0,
// 		UnbondingTokens:         0,
// 		UnbondedTokens:          0,
// 		BondedShares:            sdk.ZeroRat(),
// 		UnbondingShares:         sdk.ZeroRat(),
// 		UnbondedShares:          sdk.ZeroRat(),
// 		InflationLastTime:       0,
// 		Inflation:               sdk.NewRat(7, 100),
// 		DateLastCommissionReset: 0,
// 		PrevBondedShares:        sdk.ZeroRat(),
// 	}
// }
//
// func defaultParams() stake.Params {
// 	return stake.Params{
// 		InflationRateChange: sdk.NewRat(13, 100),
// 		InflationMax:        sdk.NewRat(20, 100),
// 		InflationMin:        sdk.NewRat(7, 100),
// 		GoalBonded:          sdk.NewRat(67, 100),
// 		MaxValidators:       100,
// 		BondDenom:           "steak",
// 	}
// }
//
// //_______________________________________________________________________________________
//
// func makeTestCodec() *wire.Codec {
// 	var cdc = wire.NewCodec()
//
// 	// Register Msgs
// 	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
// 	cdc.RegisterConcrete(bank.MsgSend{}, "test/stake/Send", nil)
// 	cdc.RegisterConcrete(bank.MsgIssue{}, "test/stake/Issue", nil)
// 	cdc.RegisterConcrete(SubmitProposalMsg{}, "simple_governance/SubmitProposalMsg", nil)
// 	cdc.RegisterConcrete(VoteMsg{}, "simple_governance/VoteMsg", nil)
//
// 	// Register AppAccount
// 	cdc.RegisterInterface((*auth.Account)(nil), nil)
// 	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/stake/Account", nil)
// 	wire.RegisterCrypto(cdc)
//
// 	return cdc
// }
//
// // hogpodge of all sorts of input required for testing
// func createTestInput(t *testing.T, isCheckTx bool, initCoins int64) (sdk.Context, auth.AccountMapper, Keeper) {
//
// 	keyStake := sdk.NewKVStoreKey("stake")
// 	keyAcc := sdk.NewKVStoreKey("acc")
// 	simpleGovKey := sdk.NewKVStoreKey("simplegov")
//
// 	db := dbm.NewMemDB()
// 	ms := store.NewCommitMultiStore(db)
// 	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
// 	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
// 	err := ms.LoadLatestVersion()
// 	require.Nil(t, err)
//
// 	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, nil, log.NewNopLogger())
// 	cdc := makeTestCodec()
// 	accountMapper := auth.NewAccountMapper(
// 		cdc,                 // amino codec
// 		keyAcc,              // target store
// 		&auth.BaseAccount{}, // prototype
// 	)
// 	ck := bank.NewKeeper(accountMapper)
// 	stakeKeeper := stake.NewKeeper(cdc, keyStake, ck, DefaultCodespace)
//
// 	// fill all the addresses with some coins
// 	for _, addr := range addrs {
// 		ck.AddCoins(ctx, addr, sdk.Coins{
// 			{"Atom", initCoins},
// 		})
// 	}
//
// 	keeper := NewKeeper(simpleGovKey, ck, stakeKeeper, DefaultCodespace)
// 	return ctx, accountMapper, keeper
// }
//
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
//
// // for incode address generation
// func testAddr(addr string) sdk.Address {
// 	res, err := sdk.Get
// 	if err != nil {
// 		panic(err)
// 	}
// 	return res
// }
