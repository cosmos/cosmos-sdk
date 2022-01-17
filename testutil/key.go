package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateCoinKey generates a new key mnemonic along with its addrress.
func GenerateCoinKey(algo keyring.SignatureAlgo, cdc codec.Codec) (sdk.AccAddress, string, error) {
	// generate a private key, with mnemonic
	info, secret, err := keyring.NewInMemory(cdc).NewMnemonic(
		"name",
		keyring.English,
		sdk.GetConfig().GetFullBIP44Path(),
		keyring.DefaultBIP39Passphrase,
		algo,
	)
	if err != nil {
		return sdk.AccAddress{}, "", err
	}
	addr, err := info.GetAddress()
	if err != nil {
		return sdk.AccAddress{}, "", err
	}
	return addr, secret, nil
}

// GenerateSaveCoinKey generates a new key mnemonic with its addrress.
// If mnemonic is provided then it's used for key generation.
// The key is saved in the keyring. The function returns error if overwrite=true and the key
// already exists.
func GenerateSaveCoinKey(
	keybase keyring.Keyring,
	keyName, mnemonic string,
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

	if exists {
		if err := keybase.Delete(keyName); err != nil {
			return sdk.AccAddress{}, "", fmt.Errorf("failed to overwrite key")
		}
	}

	var (
		record *keyring.Record
		secret string
	)

	// generate or recover a new account
	if mnemonic != "" {
		secret = mnemonic
		record, err = keybase.NewAccount(keyName, mnemonic, keyring.DefaultBIP39Passphrase, sdk.GetConfig().GetFullBIP44Path(), algo)
	} else {
		record, secret, err = keybase.NewMnemonic(keyName, keyring.English, sdk.GetConfig().GetFullBIP44Path(), keyring.DefaultBIP39Passphrase, algo)
	}
	if err != nil {
		return sdk.AccAddress{}, "", err
	}

	addr, err := record.GetAddress()
	if err != nil {
		return nil, "", err
	}
	return addr, secret, nil
}
