package utils

import (
	"encoding/binary"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryNextSequenceRecv queries the store to get the next receive sequence and
// a merkle proof.
func QueryNextSequenceRecv(
	cliCtx context.CLIContext, portID, channelID string, prove bool,
) (channeltypes.RecvResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.KeyNextSequenceRecv(portID, channelID),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return channeltypes.RecvResponse{}, err
	}

	sequence := binary.BigEndian.Uint64(res.Value)
	sequenceRes := channeltypes.NewRecvResponse(portID, channelID, sequence, res.Proof, res.Height)

	return sequenceRes, nil
}
