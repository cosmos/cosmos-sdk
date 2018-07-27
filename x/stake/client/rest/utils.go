package rest

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Contains checks if the a given query contains one of the tx types
func contains(stringSlice []string, txType string) bool {
	for _, word := range stringSlice {
		if word == txType {
			return true
		}
	}
	return false
}

// get all Validators
func getValidators(validatorKVs []sdk.KVPair, cdc *wire.Codec) ([]types.BechValidator, error) {
	// parse out the validators
	validators := make([]types.BechValidator, len(validatorKVs))
	for i, kv := range validatorKVs {

		addr := kv.Key[1:]
		validator, err := types.UnmarshalValidator(cdc, addr, kv.Value)
		if err != nil {
			return nil, err
		}

		bech32Validator, err := validator.Bech32Validator()
		if err != nil {
			return nil, err
		}
		validators[i] = bech32Validator
	}
	return validators, nil
}

// get Validator given an Account Address
func getValidator(address sdk.AccAddress, validatorKVs []sdk.KVPair, cdc *wire.Codec) (sdk.Validator, error) {
	// parse out the validators
	for i, kv := range validatorKVs {
		addr := kv.Key[1:]
		validator, err := types.UnmarshalValidator(cdc, addr, kv.Value)
		if err != nil {
			return nil, err
		}

		ownerAddress := validator.GetOwner()
		if reflect.DeepEqual(ownerAddress.Bytes(), address.Bytes()) {
			return validator, nil
		}
	}
	return nil, nil // validator Not Found
}
