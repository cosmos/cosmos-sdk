package stake

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	oldwire "github.com/tendermint/go-wire"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

func subspace(prefix []byte) (start, end []byte) {
	end = make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++
	return prefix, end
}

// custom tx codec
// TODO: use new go-wire
func makeTestCodec() *wire.Codec {

	const msgTypeSend = 0x1
	const msgTypeIssue = 0x2
	const msgTypeDeclareCandidacy = 0x3
	const msgTypeEditCandidacy = 0x4
	const msgTypeDelegate = 0x5
	const msgTypeUnbond = 0x6
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Msg }{},
		oldwire.ConcreteType{bank.SendMsg{}, msgTypeSend},
		oldwire.ConcreteType{bank.IssueMsg{}, msgTypeIssue},
		oldwire.ConcreteType{MsgDeclareCandidacy{}, msgTypeDeclareCandidacy},
		oldwire.ConcreteType{MsgEditCandidacy{}, msgTypeEditCandidacy},
		oldwire.ConcreteType{MsgDelegate{}, msgTypeDelegate},
		oldwire.ConcreteType{MsgUnbond{}, msgTypeUnbond},
	)

	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&types.AppAccount{}, accTypeApp},
	)
	cdc := wire.NewCodec()

	// cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	// bank.RegisterWire(cdc)   // Register bank.[SendMsg,IssueMsg] types.
	// crypto.RegisterWire(cdc) // Register crypto.[PubKey,PrivKey,Signature] types.
	return cdc
}

func paramsNoInflation() Params {
	return Params{
		InflationRateChange: sdk.ZeroRat,
		InflationMax:        sdk.ZeroRat,
		InflationMin:        sdk.ZeroRat,
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "fermion",
	}
}

// hogpodge of all sorts of input required for testing
func createTestInput(t *testing.T, sender sdk.Address, isCheckTx bool, initCoins int64) (sdk.Context, sdk.AccountMapper, Keeper) {
	db := dbm.NewMemDB()
	keyStake := sdk.NewKVStoreKey("stake")
	keyMain := keyStake //sdk.NewKVStoreKey("main") //TODO fix multistore

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, nil)
	cdc := makeTestCodec()
	accountMapper := auth.NewAccountMapperSealed(
		keyMain,             // target store
		&auth.BaseAccount{}, // prototype
	)
	ck := bank.NewCoinKeeper(accountMapper)
	keeper := NewKeeper(ctx, cdc, keyStake, ck)

	//params := paramsNoInflation()
	params := keeper.GetParams(ctx)

	// fill all the addresses with some coins
	for _, addr := range addrs {
		ck.AddCoins(ctx, addr, sdk.Coins{{params.BondDenom, initCoins}})
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
	return pkEd.Wrap()
}

var pks = []crypto.PubKey{
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

// for incode address generation
func testAddr(addr string) sdk.Address {
	res, err := sdk.GetAddress(addr)
	if err != nil {
		panic(err)
	}
	return res
}

// dummy addresses used for testing
var addrs = []sdk.Address{
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

// NOTE: PubKey is supposed to be the binaryBytes of the crypto.PubKey
// instead this is just being set the address here for testing purposes
func candidatesFromAddrs(ctx sdk.Context, keeper Keeper, addrs []crypto.Address, amts []int64) {
	for i := 0; i < len(amts); i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(amts[i]),
			Liabilities: sdk.NewRat(amts[i]),
			VotingPower: sdk.NewRat(amts[i]),
		}
		keeper.setCandidate(ctx, c)
	}
}

func candidatesFromAddrsEmpty(addrs []crypto.Address) (candidates Candidates) {
	for i := 0; i < len(addrs); i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.ZeroRat,
			Liabilities: sdk.ZeroRat,
			VotingPower: sdk.ZeroRat,
		}
		candidates = append(candidates, c)
	}
	return
}

//// helper function test if Candidate is changed asabci.Validator
//func testChange(t *testing.T, val Validator, chg *abci.Validator) {
//assert := assert.New(t)
//assert.Equal(val.PubKey.Bytes(), chg.PubKey)
//assert.Equal(val.VotingPower.Evaluate(), chg.Power)
//}

//// helper function test if Candidate is removed as abci.Validator
//func testRemove(t *testing.T, val Validator, chg *abci.Validator) {
//assert := assert.New(t)
//assert.Equal(val.PubKey.Bytes(), chg.PubKey)
//assert.Equal(int64(0), chg.Power)
//}
