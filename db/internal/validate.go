package validate

import dbm "github.com/cosmos/cosmos-sdk/db"

func ValidateKv(key, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	return nil
}
