package utils

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// QueryConnection queries the store to get a connection end and a merkle
// proof.
func QueryConnection(
	cliCtx context.CLIContext, connectionID string, prove bool,
) (types.ConnectionResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyConnection(connectionID),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return types.ConnectionResponse{}, err
	}

	var connection types.ConnectionEnd
	if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &connection); err != nil {
		return types.ConnectionResponse{}, err
	}

	connRes := types.NewConnectionResponse(connectionID, connection, res.Proof, res.Height)

	return connRes, nil
}

// QueryClientConnections queries the store to get the registered connection paths
// registered for a particular client and a merkle proof.
func QueryClientConnections(
	cliCtx context.CLIContext, clientID string, prove bool,
) (types.ClientConnectionsResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyClientConnections(clientID),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return types.ClientConnectionsResponse{}, err
	}

	var paths []string
	if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &paths); err != nil {
		return types.ClientConnectionsResponse{}, err
	}

	connPathsRes := types.NewClientConnectionsResponse(clientID, paths, res.Proof, res.Height)
	return connPathsRes, nil
}
