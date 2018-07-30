package rest

import (
	"fmt"
	"net/http"
	"strings"

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
		"/stake/delegators/{addr}",
		delegatorHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/delegators/{addr}/txs // Get all staking txs (i.e msgs) from a delegator
	r.HandleFunc(
		"/stake/delegators/{addr}/txs",
		delegatorTxsHandlerFn(ctx, cdc),
	).Queries("type", "{type}").Methods("GET")

	// GET /stake/delegators/{addr}/validators // Query all validators that a delegator is bonded to
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/validators",
		delegatorValidatorsHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/delegators/{addr}/validators/{addr} // Query a validator info that a delegator is bonded to
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/validators/{validatorAddr}",
		delegatorValidatorHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/delegators/{addr}/validators/{addr}

	/*
			GET /stake/delegators/{addr}/validators/{addr}/txs // Get all txs to a validator performed by a delegator
		 	GET /stake/delegators/{addr}/validators/{addr}/txs?type=bond // Get all bonding txs to a validator performed by a delegator
			GET /stake/delegators/{addr}/validators/{addr}/txs?type=unbond // Get all unbonding txs to a validator performed by a delegator
			GET /stake/delegators/{addr}/validators/{addr}/txs?type=redelegate // Get all redelegation txs to a validator performed by a delegator
	*/
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/validators/{validatorAddr}/txs",
		stakingTxsHandlerFn(ctx, cdc),
	).Queries("type", "{type}").Methods("GET")

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

// HTTP request handler to query a delegator delegations
func delegatorHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var validatorAddr sdk.AccAddress
		var output []byte // txs from a single validator
		var typesQuerySlice []string

		// read parameters
		vars := mux.Vars(r)
		bech32delegator := vars["delegatorAddr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// Get values from query

		typesQuery := r.URL.Query().Get("type")
		typesQuerySlice = strings.Split(strings.TrimSpace(typesQuery), " ")

		noQuery := len(typesQuerySlice) == 0
		isBondTx := contains(typesQuerySlice, "bond")
		isUnbondTx := contains(typesQuerySlice, "unbond")
		isRedTx := contains(typesQuerySlice, "redelegate")

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

		for i, validator := range validators {
			// get all transactions from the delegator to val and append
			output = nil
			validatorAddr = validator.Owner
			// delegation
			if noQuery || isBondTx {
				key := stake.GetDelegationKey(delegatorAddr, validatorAddr)

				res, err := ctx.QueryStore(key, storeName)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
					return
				}

				// the query will return empty if there is no data for this record
				if len(res) != 0 {
					delegation, err := types.UnmarshalDelegation(cdc, key, res)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(err.Error()))
						return
					}

					outputBond, err := cdc.MarshalJSON(delegation)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					output = append(output, outputBond...)
				}
			}
			if noQuery || isUnbondTx {

				key := stake.GetUBDKey(delegatorAddr, validatorAddr)

				res, err := ctx.QueryStore(key, storeName)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
					return
				}

				// the query will return empty if there is no data for this record
				if len(res) != 0 {
					ubd, err := types.UnmarshalUBD(cdc, key, res)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
						return
					}

					outputUnbond, err := cdc.MarshalJSON(ubd)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}

					output = append(output, outputUnbond...)
				}
			}
			if noQuery || isRedTx {
				// keyValidatorFrom := stake.GetREDsByDelFromValSrcIndexKey(delegatorAddr, validatorAddr)
				keyValidatorTo := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)

				// All redelegations from source validator (i.e unbonding)
				res, err := ctx.QueryStore(keyValidatorFrom, storeName)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't query redelegation. Error: %s", err.Error())))
					return
				}

				// the query will return empty if there is no data for this record
				if len(res) != 0 {
					red, err := types.UnmarshalRED(cdc, keyValidatorFrom, res)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
						return
					}

					outputFrom, err := cdc.MarshalJSON(red)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					output = append(output, outputFrom...)
				}

				// All redelegations to destination validator (i.e bonding)
				res, err = ctx.QueryStore(keyValidatorTo, storeName)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't query redelegation. Error: %s", err.Error())))
					return
				}

				// the query will return empty if there is no data for this record
				if len(res) != 0 {
					red, err := types.UnmarshalRED(cdc, keyValidatorTo, res)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
						return
					}

					outputTo, err := cdc.MarshalJSON(red)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
					output = append(output, outputTo...)
				}
			}
			if len(output) == 0 {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// success
			w.Write(output) // write
		}
	}
}

// HTTP request handler to query all staking txs (msgs) from a delegator
func delegatorTxsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var output []byte
		var typesQuerySlice []string
		vars := mux.Vars(r)
		bech32delegator := vars["addr"]

		delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		node, err := ctx.GetNode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't get current Node information. Error: %s", err.Error())))
			return
		}

		// Get values from query

		typesQuery := r.URL.Query().Get("type")
		typesQuerySlice = strings.Split(strings.TrimSpace(typesQuery), " ")

		query := sdk.TagDelegator + " " + delegatorAddr.String()
		noQuery := len(typesQuerySlice) == 0
		isBondTx := contains(typesQuerySlice, "bond")
		isUnbondTx := contains(typesQuerySlice, "unbond")
		isRedTx := contains(typesQuerySlice, "redelegate")
		if !isBondTx || !isUnbondTx || !isRedTx {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// TODO double check this
		if noQuery || isBondTx {
			query = query + " ANY " + sdk.TagAction
			query = query + " delegate"
		}
		if noQuery || isUnbondTx {
			query = query + " AND " + sdk.TagAction
			query = query + " begin-unbonding"

			query = query + " OR " + sdk.TagAction
			query = query + " complete-unbonding"
		}
		if noQuery || isRedTx {
			query = query + " AND " + sdk.TagAction
			query = query + " begin-redelegation"

			query = query + " OR " + sdk.TagAction
			query = query + " complete-redelegation"
		}

		page := 0
		perPage := 100
		prove := false
		res, err := node.TxSearch(query, prove, page, perPage)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		info, err := formatTxResults(cdc, res.Txs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query txs. Error: %s", err.Error())))
			return
		}
		// success
		output, err = cdc.MarshalJSON(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output) // write
	}
}

// HTTP request handler to query all delegator bonded validators
func delegatorValidatorsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var isBonded bool
		var validatorAddr sdk.AccAddress
		var bondedValidators []sdk.Validator
		var output []byte // validators
		var typesQuerySlice []string

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

		for i, validator := range validators {
			// get all transactions from the delegator to val and append
			isBonded = false
			validatorAddr = validator.Owner

			// Check if there if the delegator is bonded or redelegated to the validator

			keyDel := stake.GetDelegationKey(delegatorAddr, validatorAddr)
			keyRed := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)

			res, err := ctx.QueryStore(keyDel, storeName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
				return
			}

			if len(res) != 0 {
				isBonded = true
			}

			if !isBonded {
				res, err = ctx.QueryStore(keyRed, storeName)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
					return
				}

				if len(res) != 0 {
					isBonded = true
				}
			}

			if isBonded {
				kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
					return
				}

				// the query will return empty if there are no validators
				if len(kvs) == 0 {
					continue
				}
				validator, err := getValidator(validatorAddr, kvs, cdc)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("Couldn't get info from validator %s. Error: %s", validatorAddr.String(), err.Error())))
					return
				} else if validator == nil && err == nil {
					continue
				}
				bondedValidators = append(bondedValidators, validator)
			}
		}
		// success
		output, err = cdc.MarshalJSON(bondedValidators)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output) // write
	}
}

// HTTP request handler to query a bonded validator
func delegatorValidatorHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read parameters
		var isBonded bool
		var output []byte
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

		// Check if there if the delegator is bonded or redelegated to the validator

		keyDel := stake.GetDelegationKey(delegatorAddr, validatorAddr)
		keyRed := stake.GetREDsByDelToValDstIndexKey(delegatorAddr, validatorAddr)

		res, err := ctx.QueryStore(keyDel, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
			return
		}

		if len(res) != 0 {
			isBonded = true
		}

		if !isBonded {
			res, err = ctx.QueryStore(keyRed, storeName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't query delegation. Error: %s", err.Error())))
				return
			}

			if len(res) != 0 {
				isBonded = true
			}
		}

		if isBonded {
			kvs, err := ctx.QuerySubspace(cdc, stake.ValidatorsKey, storeName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error: %s", err.Error())))
				return
			}

			// the query will return empty if there are no validators
			if len(kvs) == 0 {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			validator, err := getValidator(validatorAddr, kvs, cdc)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Couldn't get info from validator %s. Error: %s", validatorAddr.String(), err.Error())))
				return
			} else if validator == nil && err == nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// success
			output, err = cdc.MarshalJSON(validator)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		} else {
			// delegator is not bonded to any delegator
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Write(output)
	}
}

// TODO Rename
func stakingTxsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
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

		var typesQuerySlice []string
		var output []byte // XXX not sure if this is append
		typesQuery := r.URL.Query().Get("type")
		typesQuerySlice = strings.Split(strings.TrimSpace(typesQuery), " ")

		noQuery := len(typesQuerySlice) == 0
		isBondTx := contains(typesQuerySlice, "bond")
		isUnbondTx := contains(typesQuerySlice, "unbond")
		isRedTx := contains(typesQuerySlice, "redelegate")

		w.Write(output)
	}
}

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
		valAddress, err := sdk.AccAddressFromBech32(bech32validatorAddr)
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

		// the query will return empty if there are no validators
		if len(kvs) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		validator, err := getValidator(valAddress, kvs, cdc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't query unbonding-delegation. Error: %s", err.Error())))
			return
		}
		output, err = cdc.MarshalJSON(validator)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if output == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Write(output)
	}
}
