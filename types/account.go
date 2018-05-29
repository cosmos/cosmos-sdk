package types

import (
	"encoding/hex"
	"errors"

	cmn "github.com/tendermint/tmlibs/common"
)

// Address in go-crypto style
type Address = cmn.HexBytes

// create an Address from a string
func GetAddress(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}
