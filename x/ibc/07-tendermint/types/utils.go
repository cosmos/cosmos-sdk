package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
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

// ValSetFromTmTypes casts a proto ValidatorSet to tendendermint type.
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
