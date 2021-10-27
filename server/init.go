package server

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateCoinKey(algo keyring.SignatureAlgo) (sdk.AccAddress, string, error) {
	// generate a private key, with recovery phrase
	info, secret, err := keyring.NewInMemory().NewMnemonic("name", keyring.English, sdk.GetConfig().GetFullBIP44Path(), keyring.DefaultBIP39Passphrase, algo)
	if err != nil {
		return sdk.AccAddress([]byte{}), "", err
	}
	return sdk.AccAddress(info.GetPubKey().Address()), secret, nil
}

// GenerateSaveCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateSaveCoinKey(keybase keyring.Keyring, keyName string, overwrite bool, algo keyring.SignatureAlgo) (sdk.AccAddress, string, error) {
	exists := false
	_, err := keybase.Key(keyName)
	if err == nil {
		exists = true
	}

	// ensure no overwrite
	if !overwrite && exists {
		return sdk.AccAddress([]byte{}), "", fmt.Errorf(
			"key already exists, overwrite is disabled")
	}

	// generate a private key, with recovery phrase
	if exists {
		err = keybase.Delete(keyName)
		if err != nil {
			return sdk.AccAddress([]byte{}), "", fmt.Errorf(
				"failed to overwrite key")
		}
	}

	info, secret, err := keybase.NewMnemonic(keyName, keyring.English, sdk.GetConfig().GetFullBIP44Path(), keyring.DefaultBIP39Passphrase, algo)
	if err != nil {
		return sdk.AccAddress([]byte{}), "", err
	}

	return sdk.AccAddress(info.GetPubKey().Address()), secret, nil
}
