package querier

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func validatorsToBech32(validators []types.Validator) (
	bechValidators []types.BechValidator,
	err sdk.Error,
) {

	for _, validator := range validators {
		bechValidator, err := validator.Bech32Validator()
		if err != nil {
			return nil, sdk.ErrInternal(fmt.Sprintf("could not bech32ify validator: %s", err.Error()))
		}
		bechValidators = append(bechValidators, bechValidator)
	}
	return
}
