package server

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Deprecated: GenerateCoinKey generates a new key mnemonic along with its addrress.
// Please use testutils.GenerateCoinKey instead.
func GenerateCoinKey(algo keyring.SignatureAlgo) (sdk.AccAddress, string, error) {
	// generate a private key, with mnemonic
	info, secret, err := keyring.NewInMemory().NewMnemonic(
		"name",
		keyring.English,
		sdk.GetConfig().GetFullBIP44Path(),
		keyring.DefaultBIP39Passphrase,
		algo,
	)
	if err != nil {
		return sdk.AccAddress{}, "", err
	}

	return sdk.AccAddress(info.GetPubKey().Address()), secret, nil
}

// Deprecated: GenerateSaveCoinKey generates a new key mnemonic with its addrress.
// If mnemonic is provided then it's used for key generation.
// The key is saved in the keyring. The function returns error if overwrite=true and the key
// already exists.
// Please use testutils.GenerateSaveCoinKey instead.
func GenerateSaveCoinKey(
	keybase keyring.Keyring,
	keyName string,
	overwrite bool,
	algo keyring.SignatureAlgo,
) (sdk.AccAddress, string, error) {
	exists := false
	_, err := keybase.Key(keyName)
	if err == nil {
		exists = true
	}

	// ensure no overwrite
	if !overwrite && exists {
		return sdk.AccAddress{}, "", fmt.Errorf("key already exists, overwrite is disabled")
	}

	// remove the old key by name if it exists
	if exists {
		if err := keybase.Delete(keyName); err != nil {
			return sdk.AccAddress{}, "", fmt.Errorf("failed to overwrite key")
		}
	}

	k, mnemonic, err := keybase.NewMnemonic(keyName, keyring.English, sdk.GetConfig().GetFullBIP44Path(), keyring.DefaultBIP39Passphrase, algo)
	if err != nil {
		return sdk.AccAddress{}, "", err
	}

	return sdk.AccAddress(k.GetAddress()), mnemonic, nil
}
