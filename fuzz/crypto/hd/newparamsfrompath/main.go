package newparamsfrompath

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
)

func Fuzz(data []byte) int {
	_, err := hd.NewParamsFromPath(string(data))
	if err != nil {
		return 1
	}
	return 0
}
