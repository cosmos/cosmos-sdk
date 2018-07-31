package rest

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/pkg/errors"
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
func getValidator(address sdk.ValAddress, validatorKVs []sdk.KVPair, cdc *wire.Codec) (stake.BechValidator, error) {
	// parse out the validators
	for _, kv := range validatorKVs {
		addr := kv.Key[1:]
		validator, err := types.UnmarshalValidator(cdc, addr, kv.Value)
		if err != nil {
			return stake.BechValidator{}, err
		}

		ownerAddress := validator.PubKey.Address()
		if reflect.DeepEqual(ownerAddress.Bytes(), address.Bytes()) {
			bech32Validator, err := validator.Bech32Validator()
			if err != nil {
				return stake.BechValidator{}, err
			}

			return bech32Validator, nil
		}
	}
	return stake.BechValidator{}, errors.Errorf("Couldn't find validator") // validator Not Found
}
