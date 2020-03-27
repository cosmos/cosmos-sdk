package types

import (
	"math"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/types/proto3"
	"github.com/tendermint/tendermint/version"
)

// Copied unimported test functions from tmtypes to use them here
func MakeBlockID(hash []byte, partSetSize int, partSetHash []byte) tmtypes.BlockID {
	return tmtypes.BlockID{
		Hash: hash,
		PartsHeader: tmtypes.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

// CreateTestHeader creates a mock header for testing only.
func CreateTestHeader(chainID string, height int64, timestamp time.Time, valSet *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator) Header {
	vsetHash := valSet.Hash()
	tmHeader := &tmtypes.Header{
		Version:            version.Consensus{Block: 2, App: 2},
		ChainID:            chainID,
		Height:             height,
		Time:               timestamp,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), math.MaxInt64, make([]byte, tmhash.Size)),
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
	abciHeader := tmtypes.TM2PB.Header(tmHeader)
	blockID := MakeBlockID(tmHeader.Hash(), 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet(chainID, height, 1, tmtypes.PrecommitType, valSet)
	commit, err := tmtypes.MakeCommit(blockID, height, 1, voteSet, signers, timestamp)
	if err != nil {
		panic(err)
	}

	signedHeader := SignedHeader{
		Header: &abciHeader,
		Commit: Commit{
			Height: commit.Height,
			Round:  int32(commit.Round),
			BlockID: &proto3.BlockID{
				Hash: blockID.Hash,
				PartsHeader: &proto3.PartSetHeader{
					Total: int32(blockID.PartsHeader.Total),
					Hash:  blockID.PartsHeader.Hash,
				},
			},
			Signatures: commit.Signatures,
			hash:       commit.Hash(),
			bitarray:   commit.BitArray(),
		},
	}

	valset := ValSetFromTmTypes(*valSet)
	return Header{
		SignedHeader: signedHeader,
		ValidatorSet: &valset,
	}
}
