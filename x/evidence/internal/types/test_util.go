/*
Common testing types and utility functions and methods to be used in unit and
integration testing of the evidence module.
*/
package types

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	_ Evidence = (*EquivocationEvidence)(nil)

	TestingCdc = codec.New()
)

const (
	EvidenceRouteEquivocation = "EquivocationEvidence"
	EvidenceTypeEquivocation  = "equivocation"
)

type (
	SimpleVote struct {
		Height           int64
		Round            int64
		Timestamp        time.Time
		ValidatorAddress cmn.HexBytes
		Signature        []byte
	}

	SimpleCanonicalVote struct {
		Height    int64
		Round     int64
		Timestamp time.Time
		ChainID   string
	}

	EquivocationEvidence struct {
		Power      int64
		TotalPower int64
		PubKey     crypto.PubKey
		VoteA      SimpleVote
		VoteB      SimpleVote
	}
)

func init() {
	RegisterCodec(TestingCdc)
	codec.RegisterCrypto(TestingCdc)
	TestingCdc.RegisterConcrete(EquivocationEvidence{}, "test/EquivocationEvidence", nil)
}

func (e EquivocationEvidence) Route() string            { return EvidenceRouteEquivocation }
func (e EquivocationEvidence) Type() string             { return EvidenceTypeEquivocation }
func (e EquivocationEvidence) GetHeight() int64         { return e.VoteA.Height }
func (e EquivocationEvidence) GetValidatorPower() int64 { return e.Power }
func (e EquivocationEvidence) GetTotalPower() int64     { return e.TotalPower }

func (e EquivocationEvidence) String() string {
	bz, _ := yaml.Marshal(e)
	return string(bz)
}

func (e EquivocationEvidence) GetConsensusAddress() sdk.ConsAddress {
	return sdk.ConsAddress(e.PubKey.Address())
}

func (e EquivocationEvidence) ValidateBasic() error {
	if e.VoteA.Height != e.VoteB.Height ||
		e.VoteA.Round != e.VoteB.Round {
		return fmt.Errorf("H/R/S does not match (got %v and %v)", e.VoteA, e.VoteB)
	}

	if !bytes.Equal(e.VoteA.ValidatorAddress, e.VoteB.ValidatorAddress) {
		return fmt.Errorf(
			"validator addresses do not match (got %X and %X)",
			e.VoteA.ValidatorAddress,
			e.VoteB.ValidatorAddress,
		)
	}

	return nil
}

func (e EquivocationEvidence) Hash() cmn.HexBytes {
	return tmhash.Sum(TestingCdc.MustMarshalBinaryBare(e))
}

func (v SimpleVote) SignBytes(chainID string) []byte {
	scv := SimpleCanonicalVote{
		Height:    v.Height,
		Round:     v.Round,
		Timestamp: v.Timestamp,
		ChainID:   chainID,
	}
	bz, _ := TestingCdc.MarshalBinaryLengthPrefixed(scv)
	return bz
}

func EquivocationHandler(k interface{}) Handler {
	return func(ctx sdk.Context, e Evidence) error {
		if err := e.ValidateBasic(); err != nil {
			return err
		}

		ee, ok := e.(EquivocationEvidence)
		if !ok {
			return fmt.Errorf("unexpected evidence type: %T", e)
		}
		if !ee.PubKey.VerifyBytes(ee.VoteA.SignBytes(ctx.ChainID()), ee.VoteA.Signature) {
			return errors.New("failed to verify vote A signature")
		}
		if !ee.PubKey.VerifyBytes(ee.VoteB.SignBytes(ctx.ChainID()), ee.VoteB.Signature) {
			return errors.New("failed to verify vote B signature")
		}

		// TODO: Slashing!

		return nil
	}
}
