package rest

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/pkg/errors"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// contains Checks if the a given query contains one of the tx types
func contains(stringSlice []string, txType string) bool {
	for _, word := range stringSlice {
		if word == txType {
			return true
		}
	}
	return false
}

func getDelegatorDelegations(ctx context.CoreContext, cdc *wire.Codec, delegatorAddr sdk.AccAddress, validatorAddr sdk.AccAddress) (
	outputDelegation DelegationWithoutRat, httpStatusCode int, errMsg string, err error,
) {
	delegationKey := stake.GetDelegationKey(delegatorAddr, validatorAddr)
	marshalledDelegation, err := ctx.QueryStore(delegationKey, storeName)
	if err != nil {
		return DelegationWithoutRat{}, http.StatusInternalServerError, "couldn't query delegation. Error: ", err
	}

	// the query will return empty if there is no data for this record
	if len(marshalledDelegation) != 0 {
		delegation, errUnmarshal := types.UnmarshalDelegation(cdc, delegationKey, marshalledDelegation)
		if errUnmarshal != nil {
			return DelegationWithoutRat{}, http.StatusInternalServerError, "couldn't unmarshall delegation. Error: ", errUnmarshal
		}

		outputDelegation := DelegationWithoutRat{
			DelegatorAddr: delegation.DelegatorAddr,
			ValidatorAddr: delegation.ValidatorAddr,
			Height:        delegation.Height,
			Shares:        delegation.Shares.FloatString(),
		}

		return outputDelegation, 0, "", nil
	}
	return DelegationWithoutRat{}, http.StatusNoContent, "", nil
}

func getDelegatorUndelegations(ctx context.CoreContext, cdc *wire.Codec, delegatorAddr sdk.AccAddress, validatorAddr sdk.AccAddress) (
	unbonds types.UnbondingDelegation, httpStatusCode int, errMsg string, err error,
) {
	undelegationKey := stake.GetUBDKey(delegatorAddr, validatorAddr)
	marshalledUnbondingDelegation, err := ctx.QueryStore(undelegationKey, storeName)
	if err != nil {
		return types.UnbondingDelegation{}, http.StatusInternalServerError, "couldn't query unbonding-delegation. Error: ", err
	}

	// the query will return empty if there is no data for this record
	if len(marshalledUnbondingDelegation) != 0 {
		unbondingDelegation, errUnmarshal := types.UnmarshalUBD(cdc, undelegationKey, marshalledUnbondingDelegation)
		if errUnmarshal != nil {
			return types.UnbondingDelegation{}, http.StatusInternalServerError, "couldn't unmarshall unbonding-delegation. Error: ", errUnmarshal
		}
		return unbondingDelegation, 0, "", nil
	}
	return types.UnbondingDelegation{}, http.StatusNoContent, "", nil
}

func getDelegatorRedelegations(ctx context.CoreContext, cdc *wire.Codec, delegatorAddr sdk.AccAddress, validatorAddr sdk.AccAddress) (
	regelegations types.Redelegation, httpStatusCode int, errMsg string, err error,
) {

	keyRedelegateTo := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)
	marshalledRedelegations, err := ctx.QueryStore(keyRedelegateTo, storeName)
	if err != nil {
		return types.Redelegation{}, http.StatusInternalServerError, "couldn't query redelegation. Error: ", err
	}

	if len(marshalledRedelegations) != 0 {
		redelegations, errUnmarshal := types.UnmarshalRED(cdc, keyRedelegateTo, marshalledRedelegations)
		if errUnmarshal != nil {
			return types.Redelegation{}, http.StatusInternalServerError, "couldn't unmarshall redelegations. Error: ", errUnmarshal
		}

		return redelegations, 0, "", nil
	}
	return types.Redelegation{}, http.StatusNoContent, "", nil
}

// queryTxs Queries staking txs
func queryTxs(node rpcclient.Client, cdc *wire.Codec, tag string, delegatorAddr string) ([]tx.Info, error) {
	page := 0
	perPage := 100
	prove := false
	query := fmt.Sprintf("%s='%s' AND %s='%s'", tags.Action, tag, tags.Delegator, delegatorAddr)
	res, err := node.TxSearch(query, prove, page, perPage)
	if err != nil {
		return nil, err
	}

	return tx.FormatTxResults(cdc, res.Txs)
}

// getValidators Gets all Validators
func getValidators(validatorKVs []sdk.KVPair, cdc *wire.Codec) ([]types.BechValidator, error) {
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

// getValidator Gets a validator given a ValAddress
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
	return stake.BechValidator{}, errors.Errorf("Couldn't find validator")
}

// getValidatorFromAccAdrr Gets a validator given an AccAddress
func getValidatorFromAccAdrr(address sdk.AccAddress, validatorKVs []sdk.KVPair, cdc *wire.Codec) (stake.BechValidator, error) {
	// parse out the validators
	for _, kv := range validatorKVs {
		addr := kv.Key[1:]
		validator, err := types.UnmarshalValidator(cdc, addr, kv.Value)
		if err != nil {
			return stake.BechValidator{}, err
		}

		ownerAddress := validator.PubKey.Address()
		if bytes.Equal(ownerAddress.Bytes(), address.Bytes()) {
			bech32Validator, err := validator.Bech32Validator()
			if err != nil {
				return stake.BechValidator{}, err
			}

			return bech32Validator, nil
		}
	}
	return stake.BechValidator{}, errors.Errorf("Couldn't find validator")
}
