package keyring

import (
	"errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	txsigning "github.com/cosmos/cosmos-sdk/x/tx/signing"
)

var _ Keyring = NoKeyring{}

var errNoKeyring = errors.New("no keyring configured")

type NoKeyring struct{}

func (k NoKeyring) List() ([]string, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) LookupAddressByKeyName(name string) ([]byte, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) GetPubKey(name string) (cryptotypes.PubKey, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) Sign(name string, msg []byte, signMode txsigning.SignMode) ([]byte, error) {
	return nil, errNoKeyring
}
