package rpc

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// TODO these next two functions feel kinda hacky based on their placement

// ValidatorCommand returns the validator set for a given height
func ValidatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tendermint-validator-set [height]",
		Short: "Get the full tendermint validator set at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
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

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			result, err := GetValidators(clientCtx, height, &page, &limit)
			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().Int(flags.FlagPage, rest.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, 100, "Query number of results returned per page")

	return cmd
}

// Validator output in bech32 format
type ValidatorOutput struct {
	Address          sdk.ConsAddress    `json:"address"`
	PubKey           cryptotypes.PubKey `json:"pub_key"`
	ProposerPriority int64              `json:"proposer_priority"`
	VotingPower      int64              `json:"voting_power"`
}

// Validators at a certain height output in bech32 format
type ResultValidatorsOutput struct {
	BlockHeight int64             `json:"block_height"`
	Validators  []ValidatorOutput `json:"validators"`
}

func (rvo ResultValidatorsOutput) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("block height: %d\n", rvo.BlockHeight))

	for _, val := range rvo.Validators {
		b.WriteString(
			fmt.Sprintf(`
  Address:          %s
  Pubkey:           %s
  ProposerPriority: %d
  VotingPower:      %d
		`,
				val.Address, val.PubKey, val.ProposerPriority, val.VotingPower,
			),
		)
	}

	return b.String()
}

func validatorOutput(validator *tmtypes.Validator) (ValidatorOutput, error) {
	pk, err := cryptocodec.FromTmPubKeyInterface(validator.PubKey)
	if err != nil {
		return ValidatorOutput{}, err
	}

	return ValidatorOutput{
		Address:          sdk.ConsAddress(validator.Address),
		PubKey:           pk,
		ProposerPriority: validator.ProposerPriority,
		VotingPower:      validator.VotingPower,
	}, nil
}

// GetValidators from client
func GetValidators(clientCtx client.Context, height *int64, page, limit *int) (ResultValidatorsOutput, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return ResultValidatorsOutput{}, err
	}

	validatorsRes, err := node.Validators(context.Background(), height, page, limit)
	if err != nil {
		return ResultValidatorsOutput{}, err
	}

	outputValidatorsRes := ResultValidatorsOutput{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]ValidatorOutput, len(validatorsRes.Validators)),
	}

	for i := 0; i < len(validatorsRes.Validators); i++ {
		outputValidatorsRes.Validators[i], err = validatorOutput(validatorsRes.Validators[i])
		if err != nil {
			return ResultValidatorsOutput{}, err
		}
	}

	return outputValidatorsRes, nil
}

// REST

// Validator Set at a height REST handler
func ValidatorSetRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 100)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse pagination parameters")
			return
		}

		vars := mux.Vars(r)
		height, err := strconv.ParseInt(vars["height"], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse block height")
			return
		}

		chainHeight, err := GetChainHeight(clientCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, "failed to parse chain height")
			return
		}
		if height > chainHeight {
			rest.WriteErrorResponse(w, http.StatusNotFound, "requested block height is bigger then the chain length")
			return
		}

		output, err := GetValidators(clientCtx, &height, &page, &limit)
		if rest.CheckInternalServerError(w, err) {
			return
		}
		rest.PostProcessResponse(w, clientCtx, output)
	}
}

// Latest Validator Set REST handler
func LatestValidatorSetRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 100)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse pagination parameters")
			return
		}

		output, err := GetValidators(clientCtx, nil, &page, &limit)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponse(w, clientCtx, output)
	}
}
