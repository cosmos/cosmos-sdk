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
// the value of the query, the proto encoded merkle proof, and the height of
// the Tendermint block containing the state root. The desired tendermint height
// to perform the query should be set in the client context. The query will be
// performed at one below this height (at the IAVL version) in order to obtain
// the correct merkle proof.
// Issue: https://github.com/cosmos/cosmos-sdk/issues/6567
//
// NOTE: The passed in height will only be decremented if the height set in
// the client context is greater than 2. A query for height 2 and 3 will
// return the same values (a proof and value for tendermint block height 3).
// Queries at height less than or equal to 1 is not expected to succeed.
func QueryTendermintProof(clientCtx client.Context, key []byte) ([]byte, []byte, clienttypes.Height, error) {
	height := clientCtx.Height

	// Use the IAVL height if a valid tendermint height is passed in.
	// ABCI queries at height less than or equal to 1 height may have
	// other side effects.
	if clientCtx.Height > 2 {
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
