package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
)

func abciValidatorToTmTypes(val *abci.Validator) *tmtypes.Validator {
	return &tmtypes.Validator{
		Address:     val.Address,
		VotingPower: val.Power,
	}
}

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func (valset ValidatorSet) ToTmTypes() *tmtypes.ValidatorSet {
	vals := make([]*tmtypes.Validator, len(valset.Validators))
	for i := range valset.Validators {
		vals[i] = abciValidatorToTmTypes(valset.Validators[i])
	}

	vs := tmtypes.ValidatorSet{
		Validators: vals,
		Proposer:   abciValidatorToTmTypes(valset.Proposer),
	}
	_ = vs.TotalVotingPower()
	return &vs
}

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func ValSetFromTmTypes(valset *tmtypes.ValidatorSet) *ValidatorSet {
	vals := make([]*abci.Validator, len(valset.Validators))

	for i := range valset.Validators {
		val := tmtypes.TM2PB.Validator(valset.Validators[i])
		vals[i] = &val
	}

	proposer := tmtypes.TM2PB.Validator(valset.Proposer)
	return &ValidatorSet{
		Validators:       vals,
		Proposer:         &proposer,
		totalVotingPower: valset.TotalVotingPower(),
	}
}

// ToTmTypes casts a proto SignedHeader to tendendermint type.
func (sh SignedHeader) ToTmTypes() *tmtypes.SignedHeader {
	tmHeader := &tmtypes.Header{
		Version: version.Consensus{
			Block: version.Protocol(sh.Header.Version.Block),
			App:   version.Protocol(sh.Header.Version.App),
		},
		ChainID: sh.Header.ChainID,
		Height:  sh.Header.Height,
		Time:    sh.Header.Time,
		LastBlockID: tmtypes.BlockID{
			Hash: sh.Header.LastBlockId.Hash,
			PartsHeader: tmtypes.PartSetHeader{
				Total: int(sh.Header.LastBlockId.PartsHeader.Total),
				Hash:  sh.Header.LastBlockId.PartsHeader.Hash,
			},
		},
		LastCommitHash:     sh.Header.LastCommitHash,
		DataHash:           sh.Header.DataHash,
		ValidatorsHash:     sh.Header.ValidatorsHash,
		NextValidatorsHash: sh.Header.NextValidatorsHash,
		ConsensusHash:      sh.Header.ConsensusHash,
		AppHash:            sh.Header.AppHash,
		LastResultsHash:    sh.Header.LastResultsHash,
		EvidenceHash:       sh.Header.EvidenceHash,
		ProposerAddress:    sh.Header.ProposerAddress,
	}

	return &tmtypes.SignedHeader{
		Header: tmHeader,
		Commit: sh.Commit.ToTmTypes(),
	}
}

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func SignedHeaderFromTmTypes(sh tmtypes.SignedHeader) SignedHeader {
	abciHeader := tmtypes.TM2PB.Header(sh.Header)
	return SignedHeader{
		Header: &abciHeader,
		Commit: CommitFromTmTypes(sh.Commit),
	}
}

func (bid BlockID) ToTmTypes() tmtypes.BlockID {
	return tmtypes.BlockID{
		Hash:        bid.Hash,
		PartsHeader: bid.PartsHeader.ToTmTypes(),
	}
}

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func BlockIDFromTmTypes(bid tmtypes.BlockID) *BlockID {
	return &BlockID{
		Hash:        bid.Hash,
		PartsHeader: PartSetHeaderFromTmTypes(bid.PartsHeader),
	}
}

func (ph *PartSetHeader) ToTmTypes() tmtypes.PartSetHeader {
	return tmtypes.PartSetHeader{
		Total: int(ph.Total),
		Hash:  ph.Hash,
	}
}

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func PartSetHeaderFromTmTypes(ph tmtypes.PartSetHeader) *PartSetHeader {
	return &PartSetHeader{
		Total: int32(ph.Total),
		Hash:  ph.Hash,
	}
}

// ToTmTypes
func (cs *CommitSig) ToTmTypes() tmtypes.CommitSig {
	return tmtypes.CommitSig{
		BlockIDFlag:      tmtypes.BlockIDFlag(cs.BlockIDFlag[0]),
		ValidatorAddress: cs.ValidatorAddress,
		Timestamp:        cs.Timestamp,
		Signature:        cs.Signature,
	}
}

// ToTmTypes casts a proto ToTmTypes to tendendermint type.
func (c Commit) ToTmTypes() *tmtypes.Commit {
	cSigs := make([]tmtypes.CommitSig, len(c.Signatures))
	for i := range c.Signatures {
		cSigs[i] = c.Signatures[i].ToTmTypes()
	}

	tmCommit := &tmtypes.Commit{
		Height:     c.Height,
		Round:      int(c.Round),
		BlockID:    c.BlockID.ToTmTypes(),
		Signatures: cSigs,
	}
	_ = tmCommit.Hash()
	_ = tmCommit.BitArray()
	return tmCommit
}

// ToTmTypes casts a proto ValidatorSet to tendendermint type.
func CommitFromTmTypes(c *tmtypes.Commit) Commit {
	commitSigs := make([]*CommitSig, len(c.Signatures))

	for i := range c.Signatures {
		cs := CommitSig{
			BlockIDFlag:      []byte{byte(c.Signatures[i].BlockIDFlag)},
			ValidatorAddress: c.Signatures[i].ValidatorAddress,
			Timestamp:        c.Signatures[i].Timestamp,
			Signature:        c.Signatures[i].Signature,
		}
		commitSigs[i] = &cs
	}

	return Commit{
		Height:     c.Height,
		Round:      int32(c.Round),
		BlockID:    BlockIDFromTmTypes(c.BlockID),
		Signatures: commitSigs,
		hash:       c.Hash(),
		bitArray:   c.BitArray().Bytes(),
	}
}
