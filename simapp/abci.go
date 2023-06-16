package simapp

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	VoteExtension struct {
		Hash   []byte
		Height int64
		Data   []byte
	}

	VoteExtensionHandler struct{}
)

func NewVoteExtensionHandler() VoteExtensionHandler {
	return VoteExtensionHandler{}
}

func (h VoteExtensionHandler) ExtendVote() sdk.ExtendVoteHandler {
	return func(_ sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		buf := make([]byte, 1024)

		_, err := rand.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random vote extension data: %w", err)
		}

		ve := VoteExtension{
			Hash:   req.Hash,
			Height: req.Height,
			Data:   buf,
		}

		bz, err := json.Marshal(ve)
		if err != nil {
			return nil, fmt.Errorf("failed to encode vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{
			VoteExtension: bz,
		}, nil
	}
}

func (h VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		var ve VoteExtension

		if err := json.Unmarshal(req.VoteExtension, &ve); err != nil {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		if req.Height != ve.Height || !bytes.Equal(req.Hash, ve.Hash) {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_VerifyStatus(abci.ResponseApplySnapshotChunk_ACCEPT)}, nil
	}
}
