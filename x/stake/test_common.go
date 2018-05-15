package stake

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
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6170"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6171"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6172"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6173"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6174"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6175"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6176"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6177"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6178"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6179"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6180"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6181"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6182"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6183"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6184"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6185"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6186"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6187"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6188"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6189"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6190"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6191"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6192"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6193"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6194"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6195"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6196"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6197"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6198"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6199"),
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
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB60"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB61"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB62"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB63"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB64"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB65"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB66"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB67"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB68"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB69"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB70"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB71"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB72"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB73"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB74"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB75"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB76"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB77"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB78"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB79"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB80"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB81"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB82"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB83"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB84"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB85"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB86"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB87"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB88"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB89"),
	}

	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey
)

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

// get raw genesis raw message for testing
func GetDefaultGenesisState() GenesisState {
	return GenesisState{
		Pool:   initialPool(),
		Params: defaultParams(),
	}
}

// XXX reference the common declaration of this function
func subspace(prefix []byte) (start, end []byte) {
	end = make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++
	return prefix, end
}

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
	cdc.RegisterInterface((*sdk.Account)(nil), nil)
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
func createTestInput(t *testing.T, isCheckTx bool, initCoins int64) (sdk.Context, sdk.AccountMapper, Keeper) {
	db := dbm.NewMemDB()
	keyStake := sdk.NewKVStoreKey("stake")
	keyMain := keyStake //sdk.NewKVStoreKey("main") //TODO fix multistore

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, nil, log.NewNopLogger())
	cdc := makeTestCodec()
	accountMapper := auth.NewAccountMapper(
		cdc,                 // amino codec
		keyMain,             // target store
		&auth.BaseAccount{}, // prototype
	)
	ck := bank.NewKeeper(accountMapper)
	keeper := NewKeeper(cdc, keyStake, ck, DefaultCodespace)
	keeper.setPool(ctx, initialPool())
	keeper.setParams(ctx, defaultParams())

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
