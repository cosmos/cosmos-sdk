package slashing

import (
	"encoding/hex"
	"os"
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
	"github.com/cosmos/cosmos-sdk/x/stake"
)

var (
	addrs = []sdk.Address{
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6160"),
		testAddr("A58856F0FD53BF058B4909A21AEC019107BA6161"),
	}
	pks = []crypto.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
	}
	initCoins int64 = 100
)

func createTestCodec() *wire.Codec {
	cdc := wire.NewCodec()
	sdk.RegisterWire(cdc)
	auth.RegisterWire(cdc)
	bank.RegisterWire(cdc)
	stake.RegisterWire(cdc)
	wire.RegisterCrypto(cdc)
	return cdc
}

func createTestInput(t *testing.T) (sdk.Context, bank.Keeper, stake.Keeper, Keeper) {
	keyAcc := sdk.NewKVStoreKey("acc")
	keyStake := sdk.NewKVStoreKey("stake")
	keySlashing := sdk.NewKVStoreKey("slashing")
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySlashing, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewTMLogger(os.Stdout), nil)
	cdc := createTestCodec()
	accountMapper := auth.NewAccountMapper(cdc, keyAcc, &auth.BaseAccount{})
	ck := bank.NewKeeper(accountMapper)
	sk := stake.NewKeeper(cdc, keyStake, ck, stake.DefaultCodespace)
	stake.InitGenesis(ctx, sk, stake.DefaultGenesisState())
	for _, addr := range addrs {
		ck.AddCoins(ctx, addr, sdk.Coins{
			{sk.GetParams(ctx).BondDenom, initCoins},
		})
	}
	keeper := NewKeeper(cdc, keySlashing, sk, DefaultCodespace)
	return ctx, ck, sk, keeper
}

func TestHandleDoubleSign(t *testing.T) {
	ctx, ck, sk, keeper := createTestInput(t)
	addr, val, amt := addrs[0], pks[0], int64(10)
	got := stake.NewHandler(sk)(ctx, newTestMsgDeclareCandidacy(addr, val, amt))
	require.True(t, got.IsOK())
	_ = sk.Tick(ctx)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins - amt}})
	require.Equal(t, sdk.NewRat(amt), sk.Validator(ctx, addr).GetPower())
	keeper.handleDoubleSign(ctx, 0, 0, val)
	require.Equal(t, sdk.NewRat(amt).Mul(sdk.NewRat(19).Quo(sdk.NewRat(20))), sk.Validator(ctx, addr).GetPower())
}

func TestHandleAbsentValidator(t *testing.T) {
	// TODO
}

func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

func testAddr(addr string) sdk.Address {
	res, err := sdk.GetAddress(addr)
	if err != nil {
		panic(err)
	}
	return res
}

func newTestMsgDeclareCandidacy(address sdk.Address, pubKey crypto.PubKey, amt int64) stake.MsgDeclareCandidacy {
	return stake.MsgDeclareCandidacy{
		Description:   stake.Description{},
		ValidatorAddr: address,
		PubKey:        pubKey,
		Bond:          sdk.Coin{"steak", amt},
	}
}
