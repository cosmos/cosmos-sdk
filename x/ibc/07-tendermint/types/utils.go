package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
)

func abciValidatorToTmTypes(val *abci.Validator) *tmtypes.Validator {
	return &tmtypes.Validator{
		Address:     val.Address,
		VotingPower: val.Power,
	}
}

// // ToTmTypes casts a proto ValidatorSet to tendendermint type.
// func (valset ValidatorSet) ToTmTypes() *tmtypes.ValidatorSet {
// 	vals := make([]*tmtypes.Validator, len(valset.Validators))
// 	for i := range valset.Validators {
// 		vals[i] = abciValidatorToTmTypes(valset.Validators[i])
// 	}

// 	vs := tmtypes.ValidatorSet{
// 		Validators: vals,
// 		Proposer:   abciValidatorToTmTypes(valset.Proposer),
// 	}
// 	_ = vs.TotalVotingPower()
// 	return &vs
// }

// // ValSetFromTmTypes casts a proto ValidatorSet to tendendermint type.
// func ValSetFromTmTypes(valset *tmtypes.ValidatorSet) *ValidatorSet {
// 	vals := make([]*abci.Validator, len(valset.Validators))

// 	for i := range valset.Validators {
// 		val := tmtypes.TM2PB.Validator(valset.Validators[i])
// 		vals[i] = &val
// 	}

// 	proposer := tmtypes.TM2PB.Validator(valset.Proposer)
// 	return &ValidatorSet{
// 		Validators:       vals,
// 		Proposer:         &proposer,
// 		totalVotingPower: valset.TotalVotingPower(),
// 	}
// }

func ProtoCommitToTmTypes(commit *tmproto.Commit) *tmtypes.Commit {
	commitSigs := make([]tmtypes.CommitSig, len(commit.Signatures))

	for i := range commit.Signatures {
		commitSigs[i] = tmtypes.CommitSig{
			BlockIDFlag:      commit.Signatures[i].BlockIdFlag,
			ValidatorAddress: commit.Signatures[i].ValidatorAddress,
			Timestamp:        commit.Signatures[i].Timestamp,
			Signature:        commit.Signatures[i].Signature,
		}
	}

	return &tmtypes.Commit{
		Height: commit.Height,
		Round:  commit.Round,
		BlockID: tmtypes.BlockID{
			Hash: commit.BlockID.Hash,
			PartsHeader: tmtypes.PartSetHeader{
				Total: commit.BlockID.PartsHeader.Total,
				Hash:  commit.BlockID.PartsHeader.Hash,
			},
		},
		Signatures: commitSigs,
	}
}

func ProtoHeaderToTmTypes(h *tmproto.Header) *tmtypes.Header {
	return &tmtypes.Header{
		Version: version.Consensus{
			Block: version.Protocol(h.Version.Block),
			App:   version.Protocol(h.Version.App),
		},
		ChainID: h.ChainID,
		Height:  h.Height,
		Time:    h.Time,
		LastBlockID: tmtypes.BlockID{
			Hash: h.LastBlockID.Hash,
			PartsHeader: tmtypes.PartSetHeader{
				Total: h.LastBlockID.PartsHeader.Total,
				Hash:  h.LastBlockID.PartsHeader.Hash,
			},
		},
		LastCommitHash:     h.LastCommitHash,
		DataHash:           h.DataHash,
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		AppHash:            h.AppHash,
		LastResultsHash:    h.LastResultsHash,
		EvidenceHash:       h.EvidenceHash,
		ProposerAddress:    h.ProposerAddress,
	}
}
