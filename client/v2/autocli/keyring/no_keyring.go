package keyring

import (
	"errors"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var _ Keyring = NoKeyring{}

var errNoKeyring = errors.New("no keyring configured")

type NoKeyring struct{}

func (k NoKeyring) Key(uid string) (*Record, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) KeyByAddress(address []byte) (*Record, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) GetRecordAddress(record *Record) ([]byte, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) GetRecordName(record *Record) string {
	return ""
}

func (k NoKeyring) GetRecordType(record *Record) KeyType {
	return 0
}

func (k NoKeyring) List() ([]string, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) LookupAddressByKeyName(name string) ([]byte, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) GetPubKey(name string) (cryptotypes.PubKey, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) Sign(name string, msg []byte, signMode signingv1beta1.SignMode) ([]byte, error) {
	return nil, errNoKeyring
}
