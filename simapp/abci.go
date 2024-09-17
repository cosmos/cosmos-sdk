package simapp

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// VoteExtensionHandler defines a dummy vote extension handler for SimApp.
	//
	// NOTE: This implementation is solely used for testing purposes. DO NOT use
	// in a production application!
	VoteExtensionHandler struct{}

	// VoteExtension defines the structure used to create a dummy vote extension.
	VoteExtension struct {
		Hash   []byte
		Height int64
		Data   []byte
	}
)

func NewVoteExtensionHandler() *VoteExtensionHandler {
	return &VoteExtensionHandler{}
}

func (h *VoteExtensionHandler) SetHandlers(bApp *baseapp.BaseApp) {
	bApp.SetExtendVoteHandler(h.ExtendVote())
	bApp.SetVerifyVoteExtensionHandler(h.VerifyVoteExtension())
}

func (h *VoteExtensionHandler) ExtendVote() sdk.ExtendVoteHandler {
	return func(_ sdk.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
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

		return &abci.ExtendVoteResponse{VoteExtension: bz}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
		var ve VoteExtension

		if err := json.Unmarshal(req.VoteExtension, &ve); err != nil {
			return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
		}

		switch {
		case req.Height != ve.Height:
			return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil

		case !bytes.Equal(req.Hash, ve.Hash):
			return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil

		case len(ve.Data) != 1024:
			return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
		}

		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT}, nil
	}
}
