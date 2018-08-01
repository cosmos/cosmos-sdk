package rest

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/pkg/errors"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
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

// Query staking txs
func queryTxs(node rpcclient.Client, cdc *wire.Codec, tag string, delegatorAddr string) ([]tx.Info, error) {
	page := 0
	perPage := 100
	prove := false
	query := tags.Action + "='" + tag + "' AND " + tags.Delegator + "='" + delegatorAddr + "'"
	res, err := node.TxSearch(query, prove, page, perPage)
	if err != nil {
		return nil, err
	}

	return tx.FormatTxResults(cdc, res.Txs)
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

// get Validator given a ValAddress
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

// get Validator given an AccAddress
func getValidatorFromAccAdrr(address sdk.AccAddress, validatorKVs []sdk.KVPair, cdc *wire.Codec) (stake.BechValidator, error) {
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
