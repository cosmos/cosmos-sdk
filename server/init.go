package server

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	dbm "github.com/tendermint/tendermint/libs/db"

	clkeys "github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateAccountKeyAndSecret returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateAccountKeyAndSecret() (sdk.AccAddress, string, error) {

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

// GenerateSaveAccountKeyAndSecret returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateSaveAccountKeyAndSecret(clientRoot, keyName, keyPass string, overwrite bool) (sdk.AccAddress, string, error) {

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
