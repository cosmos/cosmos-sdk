package keeper

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// dummy addresses used for testing
var (
	Addrs = []sdk.Address{
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6160", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctqyxjnwh"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6161", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctpesxxn9"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6162", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctzhrnsa6"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6163", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctr2489qg"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6164", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctytvs4pd"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6165", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ct9k6yqul"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6166", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctxcf3kjq"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6167", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ct89l9r0j"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6168", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctg6jkls2"),
		TestAddr("A58856F0FD53BF058B4909A21AEC019107BA6169", "cosmosaccaddr15ky9du8a2wlstz6fpx3p4mqpjyrm5ctf8yz2dc"),
	}

	// dummy pubkeys used for testing
	PKs = []crypto.PubKey{
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB56"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB58"),
		NewPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB59"),
	}

	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey

	addrDels = []sdk.Address{
		Addrs[0],
		Addrs[1],
	}
	addrVals = []sdk.Address{
		Addrs[2],
		Addrs[3],
		Addrs[4],
		Addrs[5],
		Addrs[6],
	}
)

//_______________________________________________________________________________________

// intended to be used with require/assert:  require.True(ValEq(...))
func ValEq(t *testing.T, exp, got types.Validator) (*testing.T, bool, string, types.Validator, types.Validator) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}

//_______________________________________________________________________________________

// create a codec used only for testing
func MakeTestCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.MsgSend{}, "test/stake/Send", nil)
	cdc.RegisterConcrete(bank.MsgIssue{}, "test/stake/Issue", nil)
	cdc.RegisterConcrete(types.MsgCreateValidator{}, "test/stake/CreateValidator", nil)
	cdc.RegisterConcrete(types.MsgEditValidator{}, "test/stake/EditValidator", nil)
	cdc.RegisterConcrete(types.MsgBeginUnbonding{}, "test/stake/BeginUnbonding", nil)
	cdc.RegisterConcrete(types.MsgCompleteUnbonding{}, "test/stake/CompleteUnbonding", nil)
	cdc.RegisterConcrete(types.MsgBeginRedelegate{}, "test/stake/BeginRedelegate", nil)
	cdc.RegisterConcrete(types.MsgCompleteRedelegate{}, "test/stake/CompleteRedelegate", nil)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/stake/Account", nil)
	wire.RegisterCrypto(cdc)

	return cdc
}

// default params without inflation
func ParamsNoInflation() types.Params {
	return types.Params{
		InflationRateChange: sdk.ZeroRat(),
		InflationMax:        sdk.ZeroRat(),
		InflationMin:        sdk.ZeroRat(),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}

// hogpodge of all sorts of input required for testing
func CreateTestInput(t *testing.T, isCheckTx bool, initCoins int64) (sdk.Context, auth.AccountMapper, PrivlegedKeeper) {

	keyStake := sdk.NewKVStoreKey("stake")
	keyAcc := sdk.NewKVStoreKey("acc")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, nil, log.NewNopLogger())
	cdc := MakeTestCodec()
	accountMapper := auth.NewAccountMapper(
		cdc,                 // amino codec
		keyAcc,              // target store
		&auth.BaseAccount{}, // prototype
	)
	ck := bank.NewKeeper(accountMapper)
	keeper := NewPrivlegedKeeper(cdc, keyStake, ck, types.DefaultCodespace)
	keeper.SetPool(ctx, types.InitialPool())
	keeper.SetNewParams(ctx, types.DefaultParams())

	// fill all the addresses with some coins
	for _, addr := range Addrs {
		ck.AddCoins(ctx, addr, sdk.Coins{
			{keeper.GetParams(ctx).BondDenom, initCoins},
		})
	}

	return ctx, accountMapper, keeper
}

func NewPubKey(pk string) (res crypto.PubKey) {
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
func TestAddr(addr string, bech string) sdk.Address {

	res, err := sdk.GetAccAddressHex(addr)
	if err != nil {
		panic(err)
	}
	bechexpected, err := sdk.Bech32ifyAcc(res)
	if err != nil {
		panic(err)
	}
	if bech != bechexpected {
		panic("Bech encoding doesn't match reference")
	}

	bechres, err := sdk.GetAccAddressBech32(bech)
	if err != nil {
		panic(err)
	}
	if bytes.Compare(bechres, res) != 0 {
		panic("Bech decode and hex decode don't match")
	}

	return res
}
