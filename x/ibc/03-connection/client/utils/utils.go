package utils

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryAllConnections returns all the connections. It _does not_ return
// any merkle proof.
func QueryAllConnections(cliCtx context.CLIContext, page, limit int) ([]types.ConnectionEnd, int64, error) {
	params := types.NewQueryAllConnectionsParams(page, limit)
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllConnections)
	res, height, err := cliCtx.QueryWithData(route, bz)
	if err != nil {
		return nil, 0, err
	}

	var connections []types.ConnectionEnd
	err = cliCtx.Codec.UnmarshalJSON(res, &connections)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal connections: %w", err)
	}
	return connections, height, nil
}

// QueryConnection queries the store to get a connection end and a merkle
// proof.
func QueryConnection(
	cliCtx context.CLIContext, connectionID string, prove bool,
) (types.ConnectionResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyConnection(connectionID),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return types.ConnectionResponse{}, err
	}

	var connection types.ConnectionEnd
	if err := cliCtx.Codec.UnmarshalBinaryBare(res.Value, &connection); err != nil {
		return types.ConnectionResponse{}, err
	}

	connRes := types.NewConnectionResponse(connectionID, connection, res.Proof, res.Height)

	return connRes, nil
}

// QueryAllClientConnectionPaths returns all the client connections paths. It
// _does not_ return any merkle proof.
func QueryAllClientConnectionPaths(cliCtx context.CLIContext, page, limit int) ([]types.ConnectionPaths, int64, error) {
	params := types.NewQueryAllConnectionsParams(page, limit)
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllClientConnections)
	res, height, err := cliCtx.QueryWithData(route, bz)
	if err != nil {
		return nil, 0, err
	}

	var connectionPaths []types.ConnectionPaths
	err = cliCtx.Codec.UnmarshalJSON(res, &connectionPaths)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal client connection paths: %w", err)
	}
	return connectionPaths, height, nil
}

// QueryClientConnections queries the store to get the registered connection paths
// registered for a particular client and a merkle proof.
func QueryClientConnections(
	cliCtx context.CLIContext, clientID string, prove bool,
) (types.ClientConnectionsResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyClientConnections(clientID),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return types.ClientConnectionsResponse{}, err
	}

	var paths []string
	if err := cliCtx.Codec.UnmarshalBinaryBare(res.Value, &paths); err != nil {
		return types.ClientConnectionsResponse{}, err
	}

	connPathsRes := types.NewClientConnectionsResponse(clientID, paths, res.Proof, res.Height)
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

// ParseProof unmarshals an cmd input argument from a JSON string to a commitment
// Proof. If the input is not a JSON, it looks for a path to the JSON file.
func ParseProof(cdc *codec.Codec, arg string) (commitmenttypes.MerkleProof, error) {
	var proof commitmenttypes.MerkleProof
	if err := cdc.UnmarshalJSON([]byte(arg), &proof); err != nil {
		// check for file path if JSON input is not provided
		contents, err := ioutil.ReadFile(arg)
		if err != nil {
			return commitmenttypes.MerkleProof{}, errors.New("neither JSON input nor path to .json file were provided")
		}
		if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
			return commitmenttypes.MerkleProof{}, errors.Wrap(err, "error unmarshalling commitment proof")
		}
	}
	return proof, nil
}
