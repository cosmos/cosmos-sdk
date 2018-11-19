package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/types"

	clkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// SimpleGenTx is a simple genesis tx
type SimpleGenTx struct {
	Addr sdk.AccAddress `json:"addr"`
}

//_____________________________________________________________________

// Generate a genesis transaction
func SimpleAppGenTx(cdc *codec.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator types.GenesisValidator, err error) {
	var addr sdk.AccAddress
	var secret string
	addr, secret, err = GenerateCoinKey()
	if err != nil {
		return
	}
	var bz []byte
	simpleGenTx := SimpleGenTx{Addr: addr}
	bz, err = cdc.MarshalJSON(simpleGenTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)
	mm := map[string]string{"secret": secret}
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)
	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}
	return
}

// create the genesis app state
func SimpleAppGenState(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	appState json.RawMessage, err error) {

	if len(appGenTxs) != 1 {
		err = errors.New("must provide a single genesis transaction")
		return
	}

	var tx SimpleGenTx
	err = cdc.UnmarshalJSON(appGenTxs[0], &tx)
	if err != nil {
		return
	}

	appState = json.RawMessage(fmt.Sprintf(`{
  "accounts": [{
    "address": "%s",
    "coins": [
      {
        "denom": "mycoin",
        "amount": "9007199254740992"
      }
    ]
  }]
}`, tx.Addr))
	return
}

//___________________________________________________________________________________________

// GenerateCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateCoinKey() (sdk.AccAddress, string, error) {

	// construct an in-memory key store
	keybase := keys.New(
		dbm.NewMemDB(),
	)

	// generate a private key, with recovery phrase
	info, secret, err := keybase.CreateMnemonic("name", keys.English, "pass", keys.Secp256k1)
	if err != nil {
		return sdk.AccAddress([]byte{}), "", err
	}
	addr := info.GetPubKey().Address()
	return sdk.AccAddress(addr), secret, nil
}

// GenerateSaveCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateSaveCoinKey(clientRoot, keyName, keyPass string, overwrite bool) (sdk.AccAddress, string, error) {

	// get the keystore from the client
	keybase, err := clkeys.GetKeyBaseFromDirWithWritePerm(clientRoot)
	if err != nil {
		return sdk.AccAddress([]byte{}), "", err
	}

	// ensure no overwrite
	if !overwrite {
		_, err := keybase.Get(keyName)
		if err == nil {
			return sdk.AccAddress([]byte{}), "", fmt.Errorf(
				"key already exists, overwrite is disabled (clientRoot: %s)", clientRoot)
		}
	}

	// generate a private key, with recovery phrase
	info, secret, err := keybase.CreateMnemonic(keyName, keys.English, keyPass, keys.Secp256k1)
	if err != nil {
		return sdk.AccAddress([]byte{}), "", err
	}
	addr := info.GetPubKey().Address()
	return sdk.AccAddress(addr), secret, nil
}
