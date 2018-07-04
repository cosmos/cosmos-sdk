package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const storeName = "stake"

func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {

	r.HandleFunc(
		"/stake/{delegator}/delegation/{validator}",
		delegationHandlerFn(ctx, cdc),
	).Methods("GET")

	r.HandleFunc(
		"/stake/{delegator}/ubd/{validator}",
		ubdHandlerFn(ctx, cdc),
	).Methods("GET")

	r.HandleFunc(
		"/stake/{delegator}/red/{validator_src}/{validator_dst}",
		redHandlerFn(ctx, cdc),
	).Methods("GET")

	r.HandleFunc(
		"/stake/validators",
		validatorsHandlerFn(ctx, cdc),
	).Methods("GET")
}

// http request handler to query a delegation
func delegationHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegator"]
		bech32validator := vars["validator"]

		delegatorAddr, err := sdk.GetAccAddressBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.GetValAddressBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key := stake.GetDelegationKey(delegatorAddr, validatorAddr)

		res, err := ctx.QueryStore(key, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this record
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		delegation, err := types.UnmarshalDelegation(cdc, key, res)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := cdc.MarshalJSON(delegation)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// http request handler to query an unbonding-delegation
func ubdHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegator"]
		bech32validator := vars["validator"]

		delegatorAddr, err := sdk.GetAccAddressBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.GetValAddressBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key := stake.GetUBDKey(delegatorAddr, validatorAddr)

		res, err := ctx.QueryStore(key, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this record
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ubd, err := types.UnmarshalUBD(cdc, key, res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
			return
		}

		output, err := cdc.MarshalJSON(ubd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// http request handler to query an redelegation
func redHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegator"]
		bech32validatorSrc := vars["validator_src"]
		bech32validatorDst := vars["validator_dst"]

		delegatorAddr, err := sdk.GetAccAddressBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorSrcAddr, err := sdk.GetValAddressBech32(bech32validatorSrc)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorDstAddr, err := sdk.GetValAddressBech32(bech32validatorDst)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key := stake.GetREDKey(delegatorAddr, validatorSrcAddr, validatorDstAddr)

		res, err := ctx.QueryStore(key, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query redelegation. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this record
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		red, err := types.UnmarshalRED(cdc, key, res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
			return
		}

		output, err := cdc.MarshalJSON(red)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// TODO move exist next to validator struct for maintainability
type StakeValidatorOutput struct {
	Owner   string `json:"owner"`   // in bech32
	PubKey  string `json:"pub_key"` // in bech32
	Revoked bool   `json:"revoked"` // has the validator been revoked from bonded status?

	PoolShares      stake.PoolShares `json:"pool_shares"`      // total shares for tokens held in the pool
	DelegatorShares sdk.Rat          `json:"delegator_shares"` // total shares issued to a validator's delegators

	Description        stake.Description `json:"description"`           // description terms for the validator
	BondHeight         int64             `json:"bond_height"`           // earliest height as a bonded validator
	BondIntraTxCounter int16             `json:"bond_intra_tx_counter"` // block-local tx index of validator change
	ProposerRewardPool sdk.Coins         `json:"proposer_reward_pool"`  // XXX reward pool collected from being the proposer

	Commission            sdk.Rat `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat `json:"commission_max"`          // XXX maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Rat `json:"commission_change_rate"`  // XXX maximum daily increase of the validator commission
	CommissionChangeToday sdk.Rat `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)

	// fee related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // total shares of a global hold pools
}

func bech32StakeValidatorOutput(validator stake.Validator) (StakeValidatorOutput, error) {
	bechOwner, err := sdk.Bech32ifyVal(validator.Owner)
	if err != nil {
		return StakeValidatorOutput{}, err
	}
	bechValPubkey, err := sdk.Bech32ifyValPub(validator.PubKey)
	if err != nil {
		return StakeValidatorOutput{}, err
	}

	return StakeValidatorOutput{
		Owner:   bechOwner,
		PubKey:  bechValPubkey,
		Revoked: validator.Revoked,

		PoolShares:      validator.PoolShares,
		DelegatorShares: validator.DelegatorShares,

		Description:        validator.Description,
		BondHeight:         validator.BondHeight,
		BondIntraTxCounter: validator.BondIntraTxCounter,
		ProposerRewardPool: validator.ProposerRewardPool,

		Commission:            validator.Commission,
		CommissionMax:         validator.CommissionMax,
		CommissionChangeRate:  validator.CommissionChangeRate,
		CommissionChangeToday: validator.CommissionChangeToday,

		PrevBondedShares: validator.PrevBondedShares,
	}, nil
}

// TODO bech32
// http request handler to query list of validators
func validatorsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query validators. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there are no validators
		if len(kvs) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// parse out the validators
		validators := make([]StakeValidatorOutput, len(kvs))
		for i, kv := range kvs {

			addr := kv.Key[1:]
			validator, err := types.UnmarshalValidator(cdc, addr, kv.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
				return
			}

			bech32Validator, err := bech32StakeValidatorOutput(validator)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			validators[i] = bech32Validator
		}

		output, err := cdc.MarshalJSON(validators)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
