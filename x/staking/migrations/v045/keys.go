package v045

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/migrations/v040"
)

func getValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(v040staking.ValidatorsKey, address.MustLengthPrefix(operatorAddr)...)
}
