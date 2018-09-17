package rpc

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"bytes"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmTypes "github.com/tendermint/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// TODO these next two functions feel kinda hacky based on their placement

//ValidatorCommand returns the validator set for a given height
func ValidatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-set [height]",
		Short: "Get the full tendermint validator set at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE:  printValidators,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(client.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().String(client.FlagChainID, "", "Chain ID of Tendermint node")
	return cmd
}

// Validator output in bech32 format
type ValidatorOutput struct {
	Address     sdk.ValAddress `json:"address"` // in bech32
	PubKey      string         `json:"pub_key"` // in bech32
	Accum       int64          `json:"accum"`
	VotingPower int64          `json:"voting_power"`
}

// Validators at a certain height output in bech32 format
type ResultValidatorsOutput struct {
	BlockHeight int64             `json:"block_height"`
	Validators  []ValidatorOutput `json:"validators"`
}

func bech32ValidatorOutput(validator *tmtypes.Validator) (ValidatorOutput, error) {
	bechValPubkey, err := sdk.Bech32ifyConsPub(validator.PubKey)
	if err != nil {
		return ValidatorOutput{}, err
	}

	return ValidatorOutput{
		Address:     sdk.ValAddress(validator.Address),
		PubKey:      bechValPubkey,
		Accum:       validator.Accum,
		VotingPower: validator.VotingPower,
	}, nil
}

func getValidators(cliCtx context.CLIContext, height *int64) ([]byte, error) {
	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	validatorsRes, err := node.Validators(height)
	if err != nil {
		return nil, err
	}

	if !cliCtx.TrustNode {
		check, err := cliCtx.Certify(validatorsRes.BlockHeight)
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(check.ValidatorsHash(), tmTypes.NewValidatorSet(validatorsRes.Validators).Hash()) {
			return nil, fmt.Errorf("got invalid validatorset")
		}
	}

	outputValidatorsRes := ResultValidatorsOutput{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]ValidatorOutput, len(validatorsRes.Validators)),
	}

	for i := 0; i < len(validatorsRes.Validators); i++ {
		outputValidatorsRes.Validators[i], err = bech32ValidatorOutput(validatorsRes.Validators[i])
		if err != nil {
			return nil, err
		}
	}

	output, err := cdc.MarshalJSON(outputValidatorsRes)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// CMD

func printValidators(cmd *cobra.Command, args []string) error {
	var height *int64
	// optional height
	if len(args) > 0 {
		h, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		if h > 0 {
			tmp := int64(h)
			height = &tmp
		}
	}

	output, err := getValidators(context.NewCLIContext(), height)
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

// REST

// Validator Set at a height REST handler
func ValidatorSetRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		height, err := strconv.ParseInt(vars["height"], 10, 64)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("ERROR: Couldn't parse block height. Assumed format is '/validatorsets/{height}'."))
			return
		}

		chainHeight, err := GetChainHeight(cliCtx)
		if height > chainHeight {
			w.WriteHeader(404)
			w.Write([]byte("ERROR: Requested block height is bigger then the chain length."))
			return
		}

		output, err := getValidators(cliCtx, &height)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// Latest Validator Set REST handler
func LatestValidatorSetRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		height, err := GetChainHeight(cliCtx)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := getValidators(cliCtx, &height)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
