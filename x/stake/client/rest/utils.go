package rest

import (
	"bytes"
	"fmt"
	"net/http"

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

// contains checks if the a given query contains one of the tx types
func contains(stringSlice []string, txType string) bool {
	for _, word := range stringSlice {
		if word == txType {
			return true
		}
	}
	return false
}

func getDelegatorValidator(cliCtx context.CLIContext, cdc *wire.Codec, delegatorAddr sdk.AccAddress, validatorAccAddr sdk.AccAddress) (
	validator types.BechValidator, httpStatusCode int, errMsg string, err error) {

	// check if the delegator is bonded or redelegated to the validator
	keyDel := stake.GetDelegationKey(delegatorAddr, validatorAccAddr)

	res, err := cliCtx.QueryStore(keyDel, storeName)
	if err != nil {
		return types.BechValidator{}, http.StatusInternalServerError, "couldn't query delegation. Error: ", err
	}

	if len(res) == 0 {
		return types.BechValidator{}, http.StatusNoContent, "", nil
	}

	kvs, err := cliCtx.QuerySubspace(stake.ValidatorsKey, storeName)
	if err != nil {
		return types.BechValidator{}, http.StatusInternalServerError, "Error: ", err
	}
	if len(kvs) == 0 {
		// the query will return empty if there are no delegations
		return types.BechValidator{}, http.StatusNoContent, "", nil
	}

	validator, errVal := getValidatorFromAccAdrr(validatorAccAddr, kvs, cdc)
	if errVal != nil {
		return types.BechValidator{}, http.StatusInternalServerError, "Couldn't get info from validator. Error: ", errVal
	}
	return validator, http.StatusOK, "", nil
}

func getDelegatorDelegations(cliCtx context.CLIContext, cdc *wire.Codec, delegatorAddr sdk.AccAddress, validatorAddr sdk.AccAddress) (
	outputDelegation DelegationWithoutRat, httpStatusCode int, errMsg string, err error) {
	delegationKey := stake.GetDelegationKey(delegatorAddr, validatorAddr)
	marshalledDelegation, err := cliCtx.QueryStore(delegationKey, storeName)
	if err != nil {
		return DelegationWithoutRat{}, http.StatusInternalServerError, "couldn't query delegation. Error: ", err
	}

	// the query will return empty if there is no data for this record
	if len(marshalledDelegation) == 0 {
		return DelegationWithoutRat{}, http.StatusNoContent, "", nil
	}

	delegation, err := types.UnmarshalDelegation(cdc, delegationKey, marshalledDelegation)
	if err != nil {
		return DelegationWithoutRat{}, http.StatusInternalServerError, "couldn't unmarshall delegation. Error: ", err
	}

	outputDelegation = DelegationWithoutRat{
		DelegatorAddr: delegation.DelegatorAddr,
		ValidatorAddr: delegation.ValidatorAddr,
		Height:        delegation.Height,
		Shares:        delegation.Shares.FloatString(),
	}

	return outputDelegation, http.StatusOK, "", nil
}

func getDelegatorUndelegations(cliCtx context.CLIContext, cdc *wire.Codec, delegatorAddr sdk.AccAddress, validatorAddr sdk.AccAddress) (
	unbonds types.UnbondingDelegation, httpStatusCode int, errMsg string, err error) {
	undelegationKey := stake.GetUBDKey(delegatorAddr, validatorAddr)
	marshalledUnbondingDelegation, err := cliCtx.QueryStore(undelegationKey, storeName)
	if err != nil {
		return types.UnbondingDelegation{}, http.StatusInternalServerError, "couldn't query unbonding-delegation. Error: ", err
	}

	// the query will return empty if there is no data for this record
	if len(marshalledUnbondingDelegation) == 0 {
		return types.UnbondingDelegation{}, http.StatusNoContent, "", nil
	}

	unbondingDelegation, err := types.UnmarshalUBD(cdc, undelegationKey, marshalledUnbondingDelegation)
	if err != nil {
		return types.UnbondingDelegation{}, http.StatusInternalServerError, "couldn't unmarshall unbonding-delegation. Error: ", err
	}
	return unbondingDelegation, http.StatusOK, "", nil
}

func getDelegatorRedelegations(cliCtx context.CLIContext, cdc *wire.Codec, delegatorAddr sdk.AccAddress, validatorAddr sdk.AccAddress) (
	regelegations types.Redelegation, httpStatusCode int, errMsg string, err error) {

	keyRedelegateTo := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)
	marshalledRedelegations, err := cliCtx.QueryStore(keyRedelegateTo, storeName)
	if err != nil {
		return types.Redelegation{}, http.StatusInternalServerError, "couldn't query redelegation. Error: ", err
	}

	if len(marshalledRedelegations) == 0 {
		return types.Redelegation{}, http.StatusNoContent, "", nil
	}

	redelegations, err := types.UnmarshalRED(cdc, keyRedelegateTo, marshalledRedelegations)
	if err != nil {
		return types.Redelegation{}, http.StatusInternalServerError, "couldn't unmarshall redelegations. Error: ", err
	}

	return redelegations, http.StatusOK, "", nil
}

// queries staking txs
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

// gets all validators
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

// gets a validator given a ValAddress
func getValidator(address sdk.AccAddress, validatorKVs []sdk.KVPair, cdc *wire.Codec) (stake.BechValidator, error) {
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

// gets a validator given an AccAddress
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

//  gets all Bech32 validators from a key
func getBech32Validators(storeName string, cliCtx context.CLIContext, cdc *wire.Codec) (
	validators []types.BechValidator, httpStatusCode int, errMsg string, err error) {
	// Get all validators using key
	kvs, err := cliCtx.QuerySubspace(stake.ValidatorsKey, storeName)
	if err != nil {
		return nil, http.StatusInternalServerError, "couldn't query validators. Error: ", err
	}

	// the query will return empty if there are no validators
	if len(kvs) == 0 {
		return nil, http.StatusNoContent, "", nil
	}

	validators, err = getValidators(kvs, cdc)
	if err != nil {
		return nil, http.StatusInternalServerError, "Error: ", err
	}
	return validators, http.StatusOK, "", nil
}
