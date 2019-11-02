package utils

import (
	"fmt"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func QueryProofs(ctx client.CLIContext, portID, channelID string, sequence uint64, queryRoute string) (types.PacketResponse, error) {
	var packetRes types.PacketResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  []byte(types.PacketCommitmentPath(portID, channelID, sequence)),
		Prove: true,
	}

	res, err := ctx.QueryABCI(req)
	if err != nil {
		return packetRes, err
	}

	fmt.Printf("res.Value: %s\n", res.Value)

	var packet types.Packet
	if err := ctx.Codec.UnmarshalJSON(res.Value, &packet); err != nil {
		return packetRes, err
	}

	return types.NewPacketResponse(portID, channelID, sequence, packet, res.Proof, res.Height), nil
}
