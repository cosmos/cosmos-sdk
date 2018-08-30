package slashing

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// TODO remove dependencies on staking (should only refer to validator set type from sdk)

var (
	pks = []crypto.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	}
	addrs = []sdk.AccAddress{
		sdk.AccAddress(pks[0].Address()),
		sdk.AccAddress(pks[1].Address()),
		sdk.AccAddress(pks[2].Address()),
	}
	initCoins sdk.Int = sdk.NewInt(200)
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
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewTMLogger(os.Stdout))
	cdc := createTestCodec()
	accountMapper := auth.NewAccountMapper(cdc, keyAcc, auth.ProtoBaseAccount)
	ck := bank.NewKeeper(accountMapper)
	sk := stake.NewKeeper(cdc, keyStake, ck, stake.DefaultCodespace)
	genesis := stake.DefaultGenesisState()

	genesis.Pool.LooseTokens = sdk.NewRat(initCoins.MulRaw(int64(len(addrs))).Int64())

	err = stake.InitGenesis(ctx, sk, genesis)
	require.Nil(t, err)

	for _, addr := range addrs {
		_, _, err = ck.AddCoins(ctx, addr, sdk.Coins{
			{sk.GetParams(ctx).BondDenom, initCoins},
		})
	}
	require.Nil(t, err)
	keeper := NewKeeper(cdc, keySlashing, sk, DefaultCodespace)
	return ctx, ck, sk, keeper
}

func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

func testAddr(addr string) sdk.AccAddress {
	res := []byte(addr)
	return res
}

func newTestMsgCreateValidator(address sdk.AccAddress, pubKey crypto.PubKey, amt sdk.Int) stake.MsgCreateValidator {
	return stake.MsgCreateValidator{
		Description:   stake.Description{},
		DelegatorAddr: address,
		ValidatorAddr: address,
		PubKey:        pubKey,
		Delegation:    sdk.Coin{"steak", amt},
	}
}
