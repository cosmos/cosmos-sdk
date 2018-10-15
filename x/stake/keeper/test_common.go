package keeper

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// dummy addresses used for testing
var (
	Addrs       = createTestAddrs(500)
	PKs         = createTestPubKeys(500)
	emptyAddr   sdk.AccAddress
	emptyPubkey crypto.PubKey

	addrDels = []sdk.AccAddress{
		Addrs[0],
		Addrs[1],
	}
	addrVals = []sdk.ValAddress{
		sdk.ValAddress(Addrs[2]),
		sdk.ValAddress(Addrs[3]),
		sdk.ValAddress(Addrs[4]),
		sdk.ValAddress(Addrs[5]),
		sdk.ValAddress(Addrs[6]),
	}
)

//_______________________________________________________________________________________

// intended to be used with require/assert:  require.True(ValEq(...))
func ValEq(t *testing.T, exp, got types.Validator) (*testing.T, bool, string, types.Validator, types.Validator) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}

//_______________________________________________________________________________________

// create a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.MsgSend{}, "test/stake/Send", nil)
	cdc.RegisterConcrete(bank.MsgIssue{}, "test/stake/Issue", nil)
	cdc.RegisterConcrete(types.MsgCreateValidator{}, "test/stake/CreateValidator", nil)
	cdc.RegisterConcrete(types.MsgEditValidator{}, "test/stake/EditValidator", nil)
	cdc.RegisterConcrete(types.MsgBeginUnbonding{}, "test/stake/BeginUnbonding", nil)
	cdc.RegisterConcrete(types.MsgBeginRedelegate{}, "test/stake/BeginRedelegate", nil)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/stake/Account", nil)
	codec.RegisterCrypto(cdc)

	return cdc
}

// default params without inflation
func ParamsNoInflation() types.Params {
	return types.Params{
		InflationRateChange: sdk.ZeroDec(),
		InflationMax:        sdk.ZeroDec(),
		InflationMin:        sdk.ZeroDec(),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}

// hogpodge of all sorts of input required for testing
func CreateTestInput(t *testing.T, isCheckTx bool, initCoins int64) (sdk.Context, auth.AccountMapper, Keeper) {

	keyStake := sdk.NewKVStoreKey("stake")
	tkeyStake := sdk.NewTransientStoreKey("transient_stake")
	keyAcc := sdk.NewKVStoreKey("acc")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tkeyStake, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, log.NewNopLogger())
	cdc := MakeTestCodec()
	accountMapper := auth.NewAccountMapper(
		cdc,                   // amino codec
		keyAcc,                // target store
		auth.ProtoBaseAccount, // prototype
	)

	ck := bank.NewBaseKeeper(accountMapper)

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	keeper := NewKeeper(cdc, keyStake, tkeyStake, ck, pk.Subspace(DefaultParamspace), types.DefaultCodespace)
	keeper.SetPool(ctx, types.InitialPool())
	keeper.SetParams(ctx, types.DefaultParams())
	keeper.InitIntraTxCounter(ctx)

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range Addrs {
		pool := keeper.GetPool(ctx)
		_, _, err := ck.AddCoins(ctx, addr, sdk.Coins{
			{keeper.BondDenom(ctx), sdk.NewInt(initCoins)},
		})
		require.Nil(t, err)
		pool.LooseTokens = pool.LooseTokens.Add(sdk.NewDec(initCoins))
		keeper.SetPool(ctx, pool)
	}

	return ctx, accountMapper, keeper
}

func NewPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	//res, err = crypto.PubKeyFromBytes(pkBytes)
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

// for incode address generation
func TestAddr(addr string, bech string) sdk.AccAddress {

	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		panic(err)
	}
	bechexpected := res.String()
	if bech != bechexpected {
		panic("Bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		panic(err)
	}
	if bytes.Compare(bechres, res) != 0 {
		panic("Bech decode and hex decode don't match")
	}

	return res
}

// nolint: unparam
func createTestAddrs(numAddrs int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHex(buffer.String())
		bech := res.String()
		addresses = append(addresses, TestAddr(buffer.String(), bech))
		buffer.Reset()
	}
	return addresses
}

// nolint: unparam
func createTestPubKeys(numPubKeys int) []crypto.PubKey {
	var publicKeys []crypto.PubKey
	var buffer bytes.Buffer

	//start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") //base pubkey string
		buffer.WriteString(numString)                                                       //adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKey(buffer.String()))
		buffer.Reset()
	}
	return publicKeys
}

//_____________________________________________________________________________________

// does a certain by-power index record exist
func ValidatorByPowerIndexExists(ctx sdk.Context, keeper Keeper, power []byte) bool {
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(power)
}

func testingUpdateValidator(keeper Keeper, ctx sdk.Context, validator types.Validator) types.Validator {
	pool := keeper.GetPool(ctx)
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator, pool)
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator, found := keeper.GetValidator(ctx, validator.OperatorAddr)
	if !found {
		panic("validator expected but not found")
	}
	return validator
}

func validatorByPowerIndexExists(k Keeper, ctx sdk.Context, power []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(power)
}
