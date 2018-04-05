package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
)

func validatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validatorset <height>",
		Short: "Get the full validator set at given height",
		RunE:  printValidators,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	// TODO: change this to false when we can
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	return cmd
}

func GetValidators(height *int64) ([]byte, error) {
	// get the node
	node, err := context.NewCoreContextFromViper().GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.Validators(height)
	if err != nil {
		return nil, err
	}

	output, err := json.MarshalIndent(res, "", "  ")
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

	output, err := GetValidators(height)
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

// REST

func ValidatorsetRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	height, err := strconv.ParseInt(vars["height"], 10, 64)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("ERROR: Couldn't parse block height. Assumed format is '/validatorsets/{height}'."))
		return
	}
	chainHeight, err := GetChainHeight()
	if height > chainHeight {
		w.WriteHeader(404)
		w.Write([]byte("ERROR: Requested block height is bigger then the chain length."))
		return
	}
	output, err := GetValidators(&height)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(output)
}

func LatestValidatorsetRequestHandler(w http.ResponseWriter, r *http.Request) {
	height, err := GetChainHeight()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	output, err := GetValidators(&height)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(output)
}
