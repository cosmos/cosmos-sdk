package utils

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryConnection queries the store to get a connection end and a merkle
// proof.
func QueryConnection(
	clientCtx client.Context, connectionID string, prove bool,
) (*types.QueryConnectionResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyConnection(connectionID),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	var connection types.ConnectionEnd
	if err := clientCtx.Codec.UnmarshalBinaryBare(res.Value, &connection); err != nil {
		return nil, err
	}

	connRes := types.NewQueryConnectionResponse(connectionID, connection, res.Proof, res.Height)

	return connRes, nil
}

// QueryClientConnections queries the store to get the registered connection paths
// registered for a particular client and a merkle proof.
func QueryClientConnections(
	clientCtx client.Context, clientID string, prove bool,
) (*types.QueryClientConnectionsResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyClientConnections(clientID),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	var paths []string
	if err := clientCtx.Codec.UnmarshalBinaryBare(res.Value, &paths); err != nil {
		return nil, err
	}

	connPathsRes := types.NewQueryClientConnectionsResponse(clientID, paths, res.Proof, res.Height)
	return connPathsRes, nil
}

// ParsePrefix unmarshals an cmd input argument from a JSON string to a commitment
// Prefix. If the input is not a JSON, it looks for a path to the JSON file.
func ParsePrefix(cdc *codec.Codec, arg string) (commitmenttypes.MerklePrefix, error) {
	var prefix commitmenttypes.MerklePrefix
	if err := cdc.UnmarshalJSON([]byte(arg), &prefix); err != nil {
		// check for file path if JSON input is not provided
		contents, err := ioutil.ReadFile(arg)
		if err != nil {
			return commitmenttypes.MerklePrefix{}, errors.New("neither JSON input nor path to .json file were provided")
		}
		if err := cdc.UnmarshalJSON(contents, &prefix); err != nil {
			return commitmenttypes.MerklePrefix{}, errors.Wrap(err, "error unmarshalling commitment prefix")
		}
	}
	return prefix, nil
}

// ParseProof unmarshals a cmd input argument from a JSON string to a commitment
// Proof. If the input is not a JSON, it looks for a path to the JSON file. It
// then marshals the commitment proof into a proto encoded byte array.
func ParseProof(cdc *codec.Codec, arg string) ([]byte, error) {
	var merkleProof commitmenttypes.MerkleProof
	if err := cdc.UnmarshalJSON([]byte(arg), &merkleProof); err != nil {
		// check for file path if JSON input is not provided
		contents, err := ioutil.ReadFile(arg)
		if err != nil {
			return nil, errors.New("neither JSON input nor path to .json file were provided")
		}
		if err := cdc.UnmarshalJSON(contents, &merkleProof); err != nil {
			return nil, fmt.Errorf("error unmarshalling commitment proof: %w", err)
		}
	}

	return cdc.MarshalBinaryBare(&merkleProof)
}
