package keeper_test

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	PKs = createTestPubKeys(500)
)

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

func NewPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	//res, err = crypto.PubKeyFromBytes(pkBytes)
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes)
	return pkEd
}

// getBaseSimappWithCustomKeeper Returns a simapp with custom StakingKeeper
// to avoid messing with the hooks.
func getBaseSimappWithCustomKeeper() (*codec.Codec, *simapp.SimApp, sdk.Context) {
	cdc := codec.New()
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	app.StakingKeeper = keeper.NewKeeper(
		simapp.NewAppCodec().Staking,
		app.GetKey(staking.StoreKey),
		app.BankKeeper,
		app.SupplyKeeper,
		app.GetSubspace(staking.ModuleName),
	)

	return cdc, app, ctx
}
