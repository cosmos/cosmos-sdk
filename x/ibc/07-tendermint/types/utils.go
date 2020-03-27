package types

import tmtypes "github.com/tendermint/tendermint/types"

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func (valset ValidatorSet) ToTmTypes() *tmtypes.ValidatorSet {
	return &tmtypes.ValidatorSet{
		Validators:       valset.Validators,
		Proposer:         valset.Proposer,
		totalVotingPower: valset.totalVotingPower,
	}
}

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func ValSetFromTmTypes(valset tmtypes.ValidatorSet) ValidatorSet {
	return ValidatorSet{
		Validators:       valset.Validators,
		Proposer:         valset.Proposer,
		totalVotingPower: valset.totalVotingPower,
	}
}

// ToTmTypes casts a proto SignedHeader to tendendermint type.
func (sh SignedHeader) ToTmTypes() *tmtypes.SignedHeader {
	return &tmtypes.SignedHeader{
		Header: &tmtypes.Header{
			Version:            sh.Header.Version,
			ChainID:            sh.Header.ChainID,
			Height:             sh.Header.Height,
			Time:               sh.Header.Time,
			LastBlockID:        sh.Header.LastBlockId,
			LastCommitHash:     sh.Header.LastCommitHash,
			DataHash:           sh.Header.DataHash,
			ValidatorsHash:     sh.Header.ValidatorsHash,
			NextValidatorsHash: sh.Header.NextValidatorsHash,
			ConsensusHash:      sh.Header.ConsensusHash,
			AppHash:            sh.Header.AppHash,
			LastResultsHash:    sh.Header.LastResult,
			EvidenceHash:       sh.Header.EvidenceHash,
			ProposerAddress:    sh.Header.ProposerAddress,
		},
		Commit: sh.Commit.ToTmTypes(),
	}
}

// ToTmTypes casts a proto ToTmTypes to tendendermint type.
func (c Commit) ToTmTypes() *tmtypes.Commit {
	return &tmtypes.Commit{
		Height:     c.Height,
		Round:      c.Round,
		BlockID:    c.BlockID,
		Signatures: c.Signatures,
		hash:       c.hash,
		bitarray:   c.bitarray,
	}
}
