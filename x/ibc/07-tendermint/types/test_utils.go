package types

import (
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/types"
	"github.com/tendermint/tendermint/proto/version"
	tmtypes "github.com/tendermint/tendermint/types"
)

const maxInt = int64(^uint(0) >> 1)

// MakeBlockID copied unimported test functions from tmtypes to use them here
func MakeBlockID(hash []byte, partSetSize int64, partSetHash []byte) tmproto.BlockID {
	return tmproto.BlockID{
		Hash: hash,
		PartsHeader: tmproto.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

// CreateTestHeader creates a mock header for testing only.
func CreateTestHeader(chainID string, height int64, timestamp time.Time, valSet *tmproto.ValidatorSet, signers []tmtypes.PrivValidator) Header {
	tmValSet, err := tmtypes.ValidatorSetFromProto(valSet)
	if err != nil {
		panic(err)
	}

	vsetHash := tmValSet.Hash()

	header := &tmproto.Header{
		Version:            version.Consensus{Block: 2, App: 2},
		ChainID:            chainID,
		Height:             height,
		Time:               timestamp,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), maxInt, make([]byte, tmhash.Size)),
		LastCommitHash:     tmhash.Sum([]byte("last_commit_hash")),
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            tmhash.Sum([]byte("app_hash")),
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    valSet.Proposer.Address,
	}

	tmHeader, err := tmtypes.HeaderFromProto(header)
	if err != nil {
		panic(err)
	}

	blockID := MakeBlockID(tmHeader.Hash(), 3, tmhash.Sum([]byte("part_set")))
	tmBlockID, err := tmtypes.BlockIDFromProto(&blockID)
	if err != nil {
		panic(err)
	}

	voteSet := tmtypes.NewVoteSet(chainID, height, 1, tmtypes.PrecommitType, tmValSet)
	commit, err := tmtypes.MakeCommit(*tmBlockID, height, 1, voteSet, signers, timestamp)
	if err != nil {
		panic(err)
	}

	signedHeader := tmproto.SignedHeader{
		Header: header,
		Commit: commit.ToProto(),
	}

	return Header{
		SignedHeader: signedHeader,
		ValidatorSet: valSet,
	}
}
