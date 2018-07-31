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

	// GET /stake/delegators/{addr} // Get all delegations (delegation, undelegation and redelegation) from a delegator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}",
		delegatorHandlerFn(ctx, cdc),
	).Methods("GET")

	// // GET /stake/delegators/{addr}/txs // Get all staking txs (i.e msgs) from a delegator
	// r.HandleFunc(
	// 	"/stake/delegators/{addr}/txs",
	// 	delegatorTxsHandlerFn(ctx, cdc),
	// ).Queries("type", "{type}").Methods("GET")

	// // GET /stake/delegators/{addr}/validators // Query all validators that a delegator is bonded to
	// r.HandleFunc(
	// 	"/stake/delegators/{delegatorAddr}/validators",
	// 	delegatorValidatorsHandlerFn(ctx, cdc),
	// ).Methods("GET")

	// GET /stake/delegators/{addr}/validators/{addr} // Query a validator info that a delegator is bonded to
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/delegations/{validatorAddr}",
		delegationHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/delegators/{addr}/validators/{addr}

	/*
			GET /stake/delegators/{addr}/validators/{addr}/txs // Get all txs to a validator performed by a delegator
		 	GET /stake/delegators/{addr}/validators/{addr}/txs?type=bond // Get all bonding txs to a validator performed by a delegator
			GET /stake/delegators/{addr}/validators/{addr}/txs?type=unbond // Get all unbonding txs to a validator performed by a delegator
			GET /stake/delegators/{addr}/validators/{addr}/txs?type=redelegate // Get all redelegation txs to a validator performed by a delegator
	*/
	// r.HandleFunc(
	// 	"/stake/delegators/{delegatorAddr}/validators/{validatorAddr}/txs",
	// 	stakingTxsHandlerFn(ctx, cdc),
	// ).Queries("type", "{type}").Methods("GET")

	// GET /stake/validators/
	r.HandleFunc(
		"/stake/validators",
		validatorsHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/validators/{addr}
	r.HandleFunc(
		"/stake/validators/{addr}",
		validatorHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/validators/{addr}/delegators
	// Don't think this is currently possible without changing keys
}

// already resolve the rational shares to not handle this in the client
type DelegationWithoutRat struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.AccAddress `json:"validator_addr"`
	Shares        string         `json:"shares"`
	Height        int64          `json:"height"`
}

// Aggregation of all delegations, unbondings and redelegations
type DelegationSummary struct {
	Delegations          []DelegationWithoutRat      `json:"delegations"`
	UnbondingDelegations []stake.UnbondingDelegation `json:"unbonding_delegations"`
	Redelegations        []stake.Redelegation        `json:"redelegations"`
}

// HTTP request handler to query a delegator delegations
func delegatorHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var validatorAddr sdk.AccAddress
		var delegationSummary = DelegationSummary{}

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// Get all validators using key
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

		validators, err := getValidators(kvs, cdc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
			return
		}

		for _, validator := range validators {
			validatorAddr = validator.Owner

			delegationKey := stake.GetDelegationKey(delegatorAddr, validatorAddr)
			marshalledDelegation, err := ctx.QueryStore(delegationKey, storeName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
				return
			}

			// the query will return empty if there is no data for this record
			if len(marshalledDelegation) != 0 {
				delegation, err := types.UnmarshalDelegation(cdc, delegationKey, marshalledDelegation)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't unmarshall unbonding-delegation. Error: %s", err.Error())))
					return
				}

				outputDelegation := DelegationWithoutRat{
					DelegatorAddr: delegation.DelegatorAddr,
					ValidatorAddr: delegation.ValidatorAddr,
					Height:        delegation.Height,
					Shares:        delegation.Shares.FloatString(),
				}

				delegationSummary.Delegations = append(delegationSummary.Delegations, outputDelegation)
			}

			undelegationKey := stake.GetUBDKey(delegatorAddr, validatorAddr)
			marshalledUnbondingDelegation, err := ctx.QueryStore(undelegationKey, storeName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
				return
			}

			// the query will return empty if there is no data for this record
			if len(marshalledUnbondingDelegation) != 0 {
				unbondingDelegation, err := types.UnmarshalUBD(cdc, undelegationKey, marshalledUnbondingDelegation)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't unmarshall unbonding-delegation. Error: %s", err.Error())))
					return
				}

				delegationSummary.UnbondingDelegations = append(delegationSummary.UnbondingDelegations, unbondingDelegation)
			}

			// only querying redelegations to a validator as this should give us already all relegations
			// if we also would put in redelegations from, we would have every redelegation double
			keyRedelegateTo := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)
			marshalledRedelegations, err := ctx.QueryStore(keyRedelegateTo, storeName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't query redelegation. Error: %s", err.Error())))
				return
			}

			if len(marshalledRedelegations) != 0 {
				redelegations, err := types.UnmarshalRED(cdc, keyRedelegateTo, marshalledRedelegations)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't unmarshall redelegations. Error: %s", err.Error())))
					return
				}

				delegationSummary.Redelegations = append(delegationSummary.Redelegations, redelegations)
			}

			output, err := cdc.MarshalJSON(delegationSummary)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// success
			w.Write(output) // write
		}
	}
}

// HTTP request handler to query all staking txs (msgs) from a delegator
// func delegatorTxsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var output []byte
// 		var typesQuerySlice []string
// 		vars := mux.Vars(r)
// 		bech32delegator := vars["addr"]

// 		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		node, err := ctx.GetNode()
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(fmt.Sprintf("Couldn't get current Node information. Error: %s", err.Error())))
// 			return
// 		}

// 		// Get values from query

// 		typesQuery := r.URL.Query().Get("type")
// 		typesQuerySlice = strings.Split(strings.TrimSpace(typesQuery), " ")

// 		query := sdk.TagDelegator + " " + delegatorAddr.String()
// 		noQuery := len(typesQuerySlice) == 0
// 		isBondTx := contains(typesQuerySlice, "bond")
// 		isUnbondTx := contains(typesQuerySlice, "unbond")
// 		isRedTx := contains(typesQuerySlice, "redelegate")
// 		if !isBondTx || !isUnbondTx || !isRedTx {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}

// 		// TODO double check this
// 		if noQuery || isBondTx {
// 			query = query + " ANY " + sdk.TagAction
// 			query = query + " delegate"
// 		}
// 		if noQuery || isUnbondTx {
// 			query = query + " AND " + sdk.TagAction
// 			query = query + " begin-unbonding"

// 			query = query + " OR " + sdk.TagAction
// 			query = query + " complete-unbonding"
// 		}
// 		if noQuery || isRedTx {
// 			query = query + " AND " + sdk.TagAction
// 			query = query + " begin-redelegation"

// 			query = query + " OR " + sdk.TagAction
// 			query = query + " complete-redelegation"
// 		}

// 		page := 0
// 		perPage := 100
// 		prove := false
// 		res, err := node.TxSearch(query, prove, page, perPage)
// 		if err != nil {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}

// 		info, err := formatTxResults(cdc, res.Txs)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(fmt.Sprintf("couldn't query txs. Error: %s", err.Error())))
// 			return
// 		}
// 		// success
// 		output, err = cdc.MarshalJSON(info)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}
// 		w.Write(output) // write
// 	}
// }

// HTTP request handler to query all delegator bonded validators
// func delegatorValidatorsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		var isBonded bool
// 		var validatorAddr sdk.AccAddress
// 		var bondedValidators []sdk.Validator
// 		var output []byte // validators
// 		var typesQuerySlice []string

// 		// read parameters
// 		vars := mux.Vars(r)
// 		bech32delegator := vars["delegatorAddr"]

// 		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		// Get all validators using key
// 		kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(fmt.Sprintf("couldn't query validators. Error: %s", err.Error())))
// 			return
// 		}

// 		// the query will return empty if there are no validators
// 		if len(kvs) == 0 {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}

// 		validators, err := getValidators(kvs, cdc)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
// 			return
// 		}

// 		for i, validator := range validators {
// 			// get all transactions from the delegator to val and append
// 			isBonded = false
// 			validatorAddr = validator.Owner

// 			// Check if there if the delegator is bonded or redelegated to the validator

// 			keyDel := stake.GetDelegationKey(delegatorAddr, validatorAddr)
// 			keyRed := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)

// 			res, err := ctx.QueryStore(keyDel, storeName)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
// 				return
// 			}

// 			if len(res) != 0 {
// 				isBonded = true
// 			}

// 			if !isBonded {
// 				res, err = ctx.QueryStore(keyRed, storeName)
// 				if err != nil {
// 					w.WriteHeader(http.StatusInternalServerError)
// 					w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
// 					return
// 				}

// 				if len(res) != 0 {
// 					isBonded = true
// 				}
// 			}

// 			if isBonded {
// 				kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
// 				if err != nil {
// 					w.WriteHeader(http.StatusInternalServerError)
// 					w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
// 					return
// 				}

// 				// the query will return empty if there are no validators
// 				if len(kvs) == 0 {
// 					continue
// 				}
// 				validator, err := getValidator(validatorAddr, kvs, cdc)
// 				if err != nil {
// 					w.WriteHeader(http.StatusInternalServerError)
// 					w.Write([]byte(fmt.Sprintf("Couldn't get info from validator %s. Error: %s", validatorAddr.String(), err.Error())))
// 					return
// 				} else if validator == nil && err == nil {
// 					continue
// 				}
// 				bondedValidators = append(bondedValidators, validator)
// 			}
// 		}
// 		// success
// 		output, err = cdc.MarshalJSON(bondedValidators)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}
// 		w.Write(output) // write
// 	}
// }

// HTTP request handler to query a bonded validator
func delegationHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]
		bech32validator := vars["validatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.AccAddressFromBech32(bech32validator)
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

		outputDelegation := DelegationWithoutRat{
			DelegatorAddr: delegation.DelegatorAddr,
			ValidatorAddr: delegation.ValidatorAddr,
			Height:        delegation.Height,
			Shares:        delegation.Shares.FloatString(),
		}

		output, err := cdc.MarshalJSON(outputDelegation)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// func delegatorValidatorHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// read parameters
// 		var isBonded bool
// 		var output []byte
// 		vars := mux.Vars(r)
// 		bech32delegator := vars["delegatorAddr"]
// 		bech32validator := vars["validatorAddr"]

// 		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		validatorAddr, err := sdk.AccAddressFromBech32(bech32validator)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		// Check if there if the delegator is bonded or redelegated to the validator

// 		keyDel := stake.GetDelegationKey(delegatorAddr, validatorAddr)
// 		keyRed := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)

// 		res, err := ctx.QueryStore(keyDel, storeName)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
// 			return
// 		}

// 		if len(res) != 0 {
// 			isBonded = true
// 		}

// 		if !isBonded {
// 			res, err = ctx.QueryStore(keyRed, storeName)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
// 				return
// 			}

// 			if len(res) != 0 {
// 				isBonded = true
// 			}
// 		}

// 		if isBonded {
// 			kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
// 				return
// 			}

// 			// the query will return empty if there are no validators
// 			if len(kvs) == 0 {
// 				w.WriteHeader(http.StatusNoContent)
// 				return
// 			}
// 			validator, err := getValidator(validatorAddr, kvs, cdc)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte(fmt.Sprintf("Couldn't get info from validator %s. Error: %s", validatorAddr.String(), err.Error())))
// 				return
// 			} else if validator == nil && err == nil {
// 				w.WriteHeader(http.StatusNoContent)
// 				return
// 			}
// 			// success
// 			output, err = cdc.MarshalJSON(validator)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte(err.Error()))
// 				return
// 			}
// 		} else {
// 			// delegator is not bonded to any delegator
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}
// 		w.Write(output)
// 	}
// }

// TODO Rename
// func stakingTxsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		// read parameters
// 		vars := mux.Vars(r)
// 		bech32delegator := vars["delegatorAddr"]
// 		bech32validator := vars["validatorAddr"]

// 		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		validatorAddr, err := sdk.AccAddressFromBech32(bech32validator)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		var typesQuerySlice []string
// 		var output []byte // XXX not sure if this is append
// 		typesQuery := r.URL.Query().Get("type")
// 		typesQuerySlice = strings.Split(strings.TrimSpace(typesQuery), " ")

// 		noQuery := len(typesQuerySlice) == 0
// 		isBondTx := contains(typesQuerySlice, "bond")
// 		isUnbondTx := contains(typesQuerySlice, "unbond")
// 		isRedTx := contains(typesQuerySlice, "redelegate")

// 		w.Write(output)
// 	}
// }

// TODO bech32
// HTTP request handler to query list of validators
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

		validators, err := getValidators(kvs, cdc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
			return
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

// HTTP request handler to query the validator information from a given validator address
func validatorHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var output []byte
		// read parameters
		vars := mux.Vars(r)
		bech32validatorAddr := vars["addr"]
		valAddress, err := sdk.ValAddressFromBech32(bech32validatorAddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
			return
		}

		kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
			return
		}

		validator, err := getValidator(valAddress, kvs, cdc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query validator. Error: %s", err.Error())))
			return
		}

		output, err = cdc.MarshalJSON(validator)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
			return
		}

		if output == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Write(output)
	}
}
