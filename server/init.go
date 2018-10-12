package server

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	dbm "github.com/tendermint/tendermint/libs/db"

	clkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// parameter names, init command
const (
	FlagName       = "name"
	FlagOverwrite  = "overwrite"
	FlagWithTxs    = "with-txs"
	FlagChainID    = "chain-id"
)

// DefaultKeyPass contains the default key password for genesis transactions
const DefaultKeyPass = "12345678"

// Storage for init command input parameters
type InitConfig struct {
	ChainID   string
	GenTxs    bool
	GenTxsDir string
	Moniker   string
	Overwrite bool
}

// Core functionality passed from the application to the server init command
type AppInit struct {

	// flags required for application init functions
	FlagsAppGenState *pflag.FlagSet

	// AppGenState creates the core parameters initialization. It takes in a
	// pubkey meant to represent the pubkey of the validator of this machine.
	AppGenState func(cdc *codec.Codec, appGenTx []json.RawMessage) (appState json.RawMessage, err error)
}

//_____________________________________________________________________

// simple default application init
var DefaultAppInit = AppInit{
	AppGenState: SimpleAppGenState,
}

// create the genesis app state
func SimpleAppGenState(cdc *codec.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	if len(appGenTxs) != 1 {
		err = errors.New("must provide a single genesis transaction")
		return
	}

	var tx auth.StdTx
	err = cdc.UnmarshalJSON(appGenTxs[0], &tx)
	if err != nil {
		return
	}
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		err = errors.New("must provide a single genesis message")
		return
	}

	msg := msgs[0].(stake.MsgCreateValidator)
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
}`, msg.ValidatorAddr))
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
	keybase, err := clkeys.GetKeyBaseFromDir(clientRoot)
	if err != nil {
		return sdk.AccAddress([]byte{}), "", err
	}

	// ensure no overwrite
	if !overwrite {
		_, err := keybase.Get(keyName)
		if err == nil {
			return sdk.AccAddress([]byte{}), "", errors.New("key already exists, overwrite is disabled")
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
