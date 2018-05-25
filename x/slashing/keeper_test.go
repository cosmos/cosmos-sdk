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
	initCoins int64 = 200
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
	genesis := stake.DefaultGenesisState()
	genesis.Pool.LooseUnbondedTokens = initCoins * int64(len(addrs))
	stake.InitGenesis(ctx, sk, genesis)
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
	addr, val, amt := addrs[0], pks[0], int64(100)
	got := stake.NewHandler(sk)(ctx, newTestMsgDeclareCandidacy(addr, val, amt))
	require.True(t, got.IsOK())
	_ = sk.Tick(ctx)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins - amt}})
	require.Equal(t, sdk.NewRat(amt), sk.Validator(ctx, addr).GetPower())
	keeper.handleDoubleSign(ctx, 0, 0, val) // double sign less than max age
	require.Equal(t, sdk.NewRat(amt).Mul(sdk.NewRat(19).Quo(sdk.NewRat(20))), sk.Validator(ctx, addr).GetPower())
	ctx = ctx.WithBlockHeader(abci.Header{Time: 300})
	keeper.handleDoubleSign(ctx, 0, 0, val) // double sign past max age
	require.Equal(t, sdk.NewRat(amt).Mul(sdk.NewRat(19).Quo(sdk.NewRat(20))), sk.Validator(ctx, addr).GetPower())
}

func TestHandleAbsentValidator(t *testing.T) {
	ctx, ck, sk, keeper := createTestInput(t)
	addr, val, amt := addrs[0], pks[0], int64(100)
	sh := stake.NewHandler(sk)
	got := sh(ctx, newTestMsgDeclareCandidacy(addr, val, amt))
	require.True(t, got.IsOK())
	_ = sk.Tick(ctx)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins - amt}})
	require.Equal(t, sdk.NewRat(amt), sk.Validator(ctx, addr).GetPower())
	info, found := keeper.getValidatorSigningInfo(ctx, val.Address())
	require.False(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.SignedBlocksCounter)
	height := int64(0)
	// 1000 blocks OK
	for ; height < 1000; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val, true)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, SignedBlocksWindow, info.SignedBlocksCounter)
	// 50 blocks missed
	for ; height < 1050; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val, false)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, SignedBlocksWindow-50, info.SignedBlocksCounter)
	// should be bonded still
	validator := sk.ValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	require.Equal(t, int64(100), pool.BondedTokens)
	// 51st block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val, false)
	info, found = keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, SignedBlocksWindow-51, info.SignedBlocksCounter)
	height++
	// should have been revoked
	validator = sk.ValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Unbonded, validator.GetStatus())
	got = sh(ctx, stake.NewMsgUnrevoke(addr))
	require.False(t, got.IsOK()) // should fail prior to jail expiration
	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(86400 * 2)})
	got = sh(ctx, stake.NewMsgUnrevoke(addr))
	require.True(t, got.IsOK()) // should succeed after jail expiration
	validator = sk.ValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	// should have been slashed
	pool = sk.GetPool(ctx)
	require.Equal(t, int64(99), pool.BondedTokens)
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
