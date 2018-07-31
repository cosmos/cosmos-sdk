package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const storeName = "stake"

func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {

	// GET /stake/delegators/{delegatorAddr} // Get all delegations (delegation, undelegation and redelegation) from a delegator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}",
		delegatorHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/delegators/{delegatorAddr}/txs // Get all staking txs (i.e msgs) from a delegator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/txs",
		delegatorTxsHandlerFn(ctx, cdc),
	).Methods("GET")

	// // GET /stake/delegators/{addr}/validators // Query all validators that a delegator is bonded to
	// r.HandleFunc(
	// 	"/stake/delegators/{delegatorAddr}/validators",
	// 	delegatorValidatorsHandlerFn(ctx, cdc),
	// ).Methods("GET")

	// GET /stake/delegators/{delegatorAddr}/delegations/{validatorAddr} // Query a delegation between a delegator and a validator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/delegations/{validatorAddr}",
		delegationHandlerFn(ctx, cdc),
	).Methods("GET")

	// GET /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} // Query all unbonding_delegations between a delegator and a validator
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}",
		unbondingDelegationsHandlerFn(ctx, cdc),
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
func delegatorTxsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var output []byte
		var typesQuerySlice []string
		vars := mux.Vars(r)
		delegatorAddr := vars["delegatorAddr"]

		_, err := sdk.AccAddressFromBech32(delegatorAddr)
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
		trimmedQuery := strings.TrimSpace(typesQuery)
		if len(trimmedQuery) != 0 {
			typesQuerySlice = strings.Split(trimmedQuery, " ")
		}

		noQuery := len(typesQuerySlice) == 0
		isBondTx := contains(typesQuerySlice, "bond")
		isUnbondTx := contains(typesQuerySlice, "unbond")
		isRedTx := contains(typesQuerySlice, "redelegate")
		var txs = []tx.TxInfo{}

		// TODO double check this
		if noQuery || isBondTx {
			foundTxs, err := queryTxs(node, cdc, string(tags.ActionDelegate), delegatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("error querying transactions. Error: %s", err.Error())))
			}
			txs = append(txs, foundTxs...)
		}
		if noQuery || isUnbondTx {
			foundTxs, err := queryTxs(node, cdc, string(tags.ActionBeginUnbonding), delegatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("error querying transactions. Error: %s", err.Error())))
			}
			txs = append(txs, foundTxs...)

			foundTxs, err = queryTxs(node, cdc, string(tags.ActionCompleteUnbonding), delegatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("error querying transactions. Error: %s", err.Error())))
			}
			txs = append(txs, foundTxs...)
		}
		if noQuery || isRedTx {
			foundTxs, err := queryTxs(node, cdc, string(tags.ActionBeginRedelegation), delegatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("error querying transactions. Error: %s", err.Error())))
			}
			txs = append(txs, foundTxs...)

			foundTxs, err = queryTxs(node, cdc, string(tags.ActionCompleteRedelegation), delegatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("error querying transactions. Error: %s", err.Error())))
			}
			txs = append(txs, foundTxs...)
		}

		// success
		output, err = cdc.MarshalJSON(txs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output) // write
	}
}

func queryTxs(node rpcclient.Client, cdc *wire.Codec, tag string, delegatorAddr string) ([]tx.TxInfo, error) {
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

// http request handler to query an unbonding-delegation
func unbondingDelegationsHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
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

		validatorAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		//TODO this seems wrong. we should query with the sdk.ValAddress and not sdk.AccAddress
		validatorAddrAcc := sdk.AccAddress(validatorAddr)

		key := stake.GetUBDKey(delegatorAddr, validatorAddrAcc)

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

		// unbondings will be a list in the future but is not yet, but we want to keep the API consistent
		ubdArray := []stake.UnbondingDelegation{ubd}

		output, err := cdc.MarshalJSON(ubdArray)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

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

		validatorAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		//TODO this seems wrong. we should query with the sdk.ValAddress and not sdk.AccAddress
		validatorAddrAcc := sdk.AccAddress(validatorAddr)

		key := stake.GetDelegationKey(delegatorAddr, validatorAddrAcc)

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
