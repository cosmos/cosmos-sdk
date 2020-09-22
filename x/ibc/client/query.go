package client

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryTendermintProof performs an ABCI query with the given key and returns
// the value of the query, the proto encoded merkle proof, and the height at
// which the proof will succeed on a tendermint verifier (one above the
// returned IAVL version height). The desired tendermint height to perform
// the query should be set in the client context. The query will be performed
// at one below this height (at the IAVL version) in order to obtain the
// correct merkle proof.
// Issue: https://github.com/cosmos/cosmos-sdk/issues/6567
func QueryTendermintProof(clientCtx client.Context, key []byte) ([]byte, []byte, clienttypes.Height, error) {
	height := clientCtx.Height

	// Use the IAVL height if a valid tendermint height is passed in.
	// ABCI queries at zero or negative height may have other side effects.
	if clientCtx.Height > 0 {
		height--
	}

	req := abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: height,
		Data:   key,
		Prove:  true,
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

	epoch := clienttypes.ParseChainID(clientCtx.ChainID)
	return res.Value, proofBz, clienttypes.NewHeight(epoch, uint64(res.Height)+1), nil
}
