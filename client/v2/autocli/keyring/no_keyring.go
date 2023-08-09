package keyring

import "errors"

var _ Keyring = NoKeyring{}

type NoKeyring struct{}

func (k NoKeyring) LookupAddressByKeyName(name string) (string, error) {
	return "", errors.New("no keyring configured")
}
