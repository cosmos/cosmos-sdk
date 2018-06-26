package stake

import (
	"bytes"
	"encoding/hex"
	"strconv"
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
)

// dummy addresses used for testing
var (
	addrs       = createTestAddrs(100)
	pks         = createTestPubKeys(100)
	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey
)

//_______________________________________________________________________________________

// intended to be used with require/assert:  require.True(ValEq(...))
func ValEq(t *testing.T, exp, got Validator) (*testing.T, bool, string, Validator, Validator) {
	return t, exp.equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}

//_______________________________________________________________________________________

func makeTestCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.MsgSend{}, "test/stake/Send", nil)
	cdc.RegisterConcrete(bank.MsgIssue{}, "test/stake/Issue", nil)
	cdc.RegisterConcrete(MsgCreateValidator{}, "test/stake/CreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "test/stake/EditValidator", nil)
	cdc.RegisterConcrete(MsgUnbond{}, "test/stake/Unbond", nil)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/stake/Account", nil)
	wire.RegisterCrypto(cdc)

	return cdc
}

func paramsNoInflation() Params {
	return Params{
		InflationRateChange: sdk.ZeroRat(),
		InflationMax:        sdk.ZeroRat(),
		InflationMin:        sdk.ZeroRat(),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}

// hogpodge of all sorts of input required for testing
func createTestInput(t *testing.T, isCheckTx bool, initCoins sdk.Int) (sdk.Context, auth.AccountMapper, Keeper) {

	keyStake := sdk.NewKVStoreKey("stake")
	keyAcc := sdk.NewKVStoreKey("acc")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, log.NewNopLogger())
	cdc := makeTestCodec()
	accountMapper := auth.NewAccountMapper(
		cdc,                 // amino codec
		keyAcc,              // target store
		&auth.BaseAccount{}, // prototype
	)
	ck := bank.NewKeeper(accountMapper)
	keeper := NewKeeper(cdc, keyStake, ck, DefaultCodespace)
	keeper.setPool(ctx, InitialPool())
	keeper.setNewParams(ctx, DefaultParams())

	// fill all the addresses with some coins
	for _, addr := range addrs {
		ck.AddCoins(ctx, addr, sdk.Coins{
			{keeper.GetParams(ctx).BondDenom, initCoins},
		})
	}

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
func testAddr(addr string, bech string) sdk.Address {

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

func createTestAddrs(numAddrs int) []sdk.Address {
	var addresses []sdk.Address
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, _ := sdk.GetAccAddressHex(buffer.String())
		bech, _ := sdk.Bech32ifyAcc(res)
		addresses = append(addresses, testAddr(buffer.String(), bech))
		buffer.Reset()
	}
	return addresses
}

func createTestPubKeys(numPubKeys int) []crypto.PubKey {
	var publicKeys []crypto.PubKey
	var buffer bytes.Buffer

	//start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") //base pubkey string
		buffer.WriteString(numString)                                                       //adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, newPubKey(buffer.String()))
		buffer.Reset()
	}
	return publicKeys
}
