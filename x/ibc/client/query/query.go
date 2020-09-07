package query

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	abci "github.com/tendermint/tendermint/abci/types"
)

// QueryTendermintProof performs an ABCI query with the given key and returns the value
// of the query, the proto encoded merkle proof for the query and the height
// at which the proof will succeed on a tendermint verifier (one above the
// returned IAVL version height).
// Issue: https://github.com/cosmos/cosmos-sdk/issues/6567
func QueryTendermintProof(clientCtx client.Context, key []byte) ([]byte, []byte, clienttypes.Height, error) {
	req := abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", host.StoreKey),
		Data:  key,
		Prove: true,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, nil, clienttypes.Height{}, err
	}

	merkleProof := commitmenttypes.MerkleProof{
		Proof: res.ProofOps,
	}

	cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

	proofBz, err := cdc.MarshalBinaryBare(&merkleProof)
	if err != nil {
		return nil, nil, clienttypes.Height{}, err
	}

	// TODO: retrieve epoch number from chain-id
	return res.Value, proofBz, clienttypes.NewHeight(0, uint64(res.Height)+1), nil
}
