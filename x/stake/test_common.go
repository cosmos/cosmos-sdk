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

var (
	// dummy addresses used for testing
	addrs = createTestAddrs(40)
	// dummy pubkeys used for testing
	pks = createTestPubKeys(40)

	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey
)

func createTestAddrs(numAddrs int) []sdk.Address {
	var addresses []sdk.Address
	var buffer bytes.Buffer

	//start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 10; i < numAddrs; i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA61") //base address string
		buffer.WriteString(numString)                                //adding on final two digits to make addresses unique
		addresses = append(addresses, testAddr(buffer.String()))
		buffer.Reset()
	}
	return addresses
}

func createTestPubKeys(numPubKeys int) []crypto.PubKey {
	var publicKeys []crypto.PubKey
	var buffer bytes.Buffer

	//start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 10; i < numPubKeys; i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB") //base pubkey string
		buffer.WriteString(numString)                                                        //adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, newPubKey(buffer.String()))
		buffer.Reset()
	}
	return publicKeys
}

func validatorsEqual(b1, b2 Validator) bool {
	return bytes.Equal(b1.Address, b2.Address) &&
		b1.PubKey.Equals(b2.PubKey) &&
		b1.Power.Equal(b2.Power) &&
		b1.Height == b2.Height &&
		b1.Counter == b2.Counter
}

func candidatesEqual(c1, c2 Candidate) bool {
	return c1.Status == c2.Status &&
		c1.PubKey.Equals(c2.PubKey) &&
		bytes.Equal(c1.Address, c2.Address) &&
		c1.Assets.Equal(c2.Assets) &&
		c1.Liabilities.Equal(c2.Liabilities) &&
		c1.Description == c2.Description
}

func bondsEqual(b1, b2 DelegatorBond) bool {
	return bytes.Equal(b1.DelegatorAddr, b2.DelegatorAddr) &&
		bytes.Equal(b1.CandidateAddr, b2.CandidateAddr) &&
		b1.Height == b2.Height &&
		b1.Shares.Equal(b2.Shares)
}

// default params for testing
func defaultParams() Params {
	return Params{
		InflationRateChange: sdk.NewRat(13, 100),
		InflationMax:        sdk.NewRat(20, 100),
		InflationMin:        sdk.NewRat(7, 100),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}

// initial pool for testing
func initialPool() Pool {
	return Pool{
		TotalSupply:       0,
		BondedShares:      sdk.ZeroRat(),
		UnbondedShares:    sdk.ZeroRat(),
		BondedPool:        0,
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
}

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
	cdc.RegisterConcrete(MsgDeclareCandidacy{}, "test/stake/DeclareCandidacy", nil)
	cdc.RegisterConcrete(MsgEditCandidacy{}, "test/stake/EditCandidacy", nil)
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
func createTestInput(t *testing.T, isCheckTx bool, initCoins int64) (sdk.Context, auth.AccountMapper, Keeper) {

	keyStake := sdk.NewKVStoreKey("stake")
	keyAcc := sdk.NewKVStoreKey("acc")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, nil, log.NewNopLogger())
	cdc := makeTestCodec()
	accountMapper := auth.NewAccountMapper(
		cdc,                 // amino codec
		keyAcc,              // target store
		&auth.BaseAccount{}, // prototype
	)
	ck := bank.NewKeeper(accountMapper)
	keeper := NewKeeper(cdc, keyStake, ck, DefaultCodespace)
	keeper.setPool(ctx, initialPool())
	keeper.setNewParams(ctx, defaultParams())

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
func testAddr(addr string) sdk.Address {
	res, err := sdk.GetAddress(addr)
	if err != nil {
		panic(err)
	}
	return res
}
