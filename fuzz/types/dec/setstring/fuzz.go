package setstring

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Fuzz(b []byte) int {
	dec, err := sdk.NewDecFromStr(string(b))
	if err != nil {
		return 0
	}
	if !dec.IsZero() {
		return 1
	}
	switch s := string(b); {
	case strings.TrimLeft(s, "-+0") == "":
		return 1
	case strings.TrimRight(strings.TrimLeft(s, "-+0"), "0") == ".":
		return 1
	default:
		panic("no error yet is zero")
	}
}
