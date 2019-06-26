package tendermint

import (
	"bytes"
	"sort"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// reimplementing tmtypes.MockPV to make it marshallable
type mockPV struct {
	PrivKey crypto.PrivKey
}

var _ tmtypes.PrivValidator = (*mockPV)(nil)

func newMockPV() *mockPV {
	return &mockPV{ed25519.GenPrivKey()}
}

func (pv *mockPV) GetAddress() tmtypes.Address {
	return pv.PrivKey.PubKey().Address()
}

func (pv *mockPV) GetPubKey() crypto.PubKey {
	return pv.PrivKey.PubKey()
}

func (pv *mockPV) SignVote(chainID string, vote *tmtypes.Vote) error {
	signBytes := vote.SignBytes(chainID)
	sig, err := pv.PrivKey.Sign(signBytes)
	if err != nil {
		return err
	}
	vote.Signature = sig
	return nil
}

func (pv *mockPV) SignProposal(string, *tmtypes.Proposal) error {
	panic("not needed")
}

// MockValset
type MockValidator struct {
	MockPV *mockPV
	Power  sdk.Dec
}

func NewMockValidator(power sdk.Dec) MockValidator {
	return MockValidator{
		MockPV: newMockPV(),
		Power:  power,
	}
}

func (val MockValidator) GetOperator() sdk.ValAddress {
	return sdk.ValAddress(val.MockPV.GetAddress())
}

func (val MockValidator) GetConsAddr() sdk.ConsAddress {
	return sdk.GetConsAddress(val.MockPV.GetPubKey())
}

func (val MockValidator) GetConsPubKey() crypto.PubKey {
	return val.MockPV.GetPubKey()
}

func (val MockValidator) GetPower() sdk.Dec {
	return val.Power
}

func (val MockValidator) Validator() *tmtypes.Validator {
	return tmtypes.NewValidator(
		val.GetConsPubKey(),
		val.GetPower().RoundInt64(),
	)
}

type MockValidators []MockValidator

var _ sort.Interface = MockValidators{}

func NewMockValidators(num int, power int64) MockValidators {
	res := make(MockValidators, num)
	for i := range res {
		res[i] = NewMockValidator(sdk.NewDec(power))
	}

	sort.Sort(res)

	return res
}

func (vals MockValidators) Len() int {
	return len(vals)
}

func (vals MockValidators) Less(i, j int) bool {
	return bytes.Compare([]byte(vals[i].GetConsAddr()), []byte(vals[j].GetConsAddr())) == -1
}

func (vals MockValidators) Swap(i, j int) {
	it := vals[j]
	vals[j] = vals[i]
	vals[i] = it
}

func (vals MockValidators) TotalPower() sdk.Dec {
	res := sdk.ZeroDec()
	for _, val := range vals {
		res = res.Add(val.Power)
	}
	return res
}

func (vals MockValidators) Sign(header tmtypes.Header) tmtypes.SignedHeader {

	precommits := make([]*tmtypes.CommitSig, len(vals))
	for i, val := range vals {
		vote := &tmtypes.Vote{
			BlockID: tmtypes.BlockID{
				Hash: header.Hash(),
			},
			ValidatorAddress: val.MockPV.GetAddress(),
			ValidatorIndex:   i,
			Height:           header.Height,
			Type:             tmtypes.PrecommitType,
		}
		val.MockPV.SignVote(chainid, vote)
		precommits[i] = vote.CommitSig()
	}

	return tmtypes.SignedHeader{
		Header: &header,
		Commit: &tmtypes.Commit{
			BlockID: tmtypes.BlockID{
				Hash: header.Hash(),
			},
			Precommits: precommits,
		},
	}
}

// Mutate valset
func (vals MockValidators) Mutate() MockValidators {
	num := len(vals) / 20 // 5% change each block

	res := make(MockValidators, len(vals))

	for i := 0; i < len(vals)-num; i++ {
		res[i] = vals[num:][i]
	}

	for i := len(vals) - num; i < len(vals); i++ {
		res[i] = NewMockValidator(vals[0].Power)
	}

	sort.Sort(res)

	for i, val := range vals {
		if val != res[i] {
			return res
		}
	}

	panic("not mutated")
}

func (vals MockValidators) ValidatorSet() *tmtypes.ValidatorSet {
	tmvals := make([]*tmtypes.Validator, len(vals))

	for i, val := range vals {
		tmvals[i] = val.Validator()
	}

	return tmtypes.NewValidatorSet(tmvals)
}
