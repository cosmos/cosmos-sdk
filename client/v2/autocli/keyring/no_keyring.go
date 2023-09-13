package keyring

import "errors"

var _ Keyring = NoKeyring{}

type NoKeyring struct{}

func (k NoKeyring) LookupAddressByKeyName(name string) ([]byte, error) {
	return nil, errors.New("no keyring configured")
}
