package client

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default gas parameters
var (
	DefaultGasSetting            = GasSetting{false, flags.DefaultGasLimit}
	DefaultGasAdjustment float64 = 1
	DefaultGasPrices             = sdk.NewDecCoins()
)

// GasSetting encapsulates the possible values passed through the --gas flag.
type GasSetting struct {
	Simulate bool
	Gas      uint64
}

func (v *GasSetting) String() string {
	if v.Simulate {
		return flags.GasFlagAuto
	}

	return strconv.FormatUint(v.Gas, 10)
}

// ParseGasSetting parses a string gas value. The value may either be 'auto',
// which indicates a transaction should be executed in simulate mode to
// automatically find a sufficient gas value, or a string integer. It returns an
// error if a string integer is provided which cannot be parsed.
func ParseGasSetting(gasStr string) (GasSetting, error) {
	switch gasStr {
	case "":
		return DefaultGasSetting, nil

	case flags.GasFlagAuto:
		return GasSetting{true, 0}, nil

	default:
		gas, err := strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			return GasSetting{}, fmt.Errorf("gas must be either integer or %s", flags.GasFlagAuto)
		}

		return GasSetting{false, gas}, nil
	}
}
