package keeper

import (
	"github.com/tendermint/tendermint/crypto/ed25519"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type testInput struct {
	cdc    *codec.Codec
	ctx    sdk.Context
	ak     auth.AccountKeeper
	pk     params.Keeper
	bk     bank.Keeper
	dk     Keeper
	router sdk.Router
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	//TODO create test input
	return testInput{}
}

var (
	senderPub     = ed25519.GenPrivKey().PubKey()
	recipientPub  = ed25519.GenPrivKey().PubKey()
	senderAddr    = sdk.AccAddress(senderPub.Address())
	recipientAddr = sdk.AccAddress(recipientPub.Address())
)
