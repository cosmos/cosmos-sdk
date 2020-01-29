/*
Common testing types and utility functions and methods to be used in unit and
integration testing of the evidence module.
*/
// DONTCOVER
package types

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"

	"gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

var (
	_ exported.Evidence = (*TestEquivocationEvidence)(nil)

	TestingCdc = codec.New()
)

const (
	TestEvidenceRouteEquivocation = "TestEquivocationEvidence"
	TestEvidenceTypeEquivocation  = "equivocation"
)

type (
	TestVote struct {
		Height           int64
		Round            int64
		Timestamp        time.Time
		ValidatorAddress tmbytes.HexBytes
		Signature        []byte
	}

	TestCanonicalVote struct {
		Height    int64
		Round     int64
		Timestamp time.Time
		ChainID   string
	}

	TestEquivocationEvidence struct {
		Power      int64
		TotalPower int64
		PubKey     crypto.PubKey
		VoteA      TestVote
		VoteB      TestVote
	}
)

func init() {
	RegisterCodec(TestingCdc)
	codec.RegisterCrypto(TestingCdc)
	TestingCdc.RegisterConcrete(TestEquivocationEvidence{}, "test/TestEquivocationEvidence", nil)
}

func (e TestEquivocationEvidence) Route() string            { return TestEvidenceRouteEquivocation }
func (e TestEquivocationEvidence) Type() string             { return TestEvidenceTypeEquivocation }
func (e TestEquivocationEvidence) GetHeight() int64         { return e.VoteA.Height }
func (e TestEquivocationEvidence) GetValidatorPower() int64 { return e.Power }
func (e TestEquivocationEvidence) GetTotalPower() int64     { return e.TotalPower }

func (e TestEquivocationEvidence) String() string {
	bz, _ := yaml.Marshal(e)
	return string(bz)
}

func (e TestEquivocationEvidence) GetConsensusAddress() sdk.ConsAddress {
	return sdk.ConsAddress(e.PubKey.Address())
}

func (e TestEquivocationEvidence) ValidateBasic() error {
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

func (e TestEquivocationEvidence) Hash() tmbytes.HexBytes {
	return tmhash.Sum(TestingCdc.MustMarshalBinaryBare(e))
}

func (v TestVote) SignBytes(chainID string) []byte {
	scv := TestCanonicalVote{
		Height:    v.Height,
		Round:     v.Round,
		Timestamp: v.Timestamp,
		ChainID:   chainID,
	}
	bz, _ := TestingCdc.MarshalBinaryLengthPrefixed(scv)
	return bz
}

func TestEquivocationHandler(k interface{}) Handler {
	return func(ctx sdk.Context, e exported.Evidence) error {
		if err := e.ValidateBasic(); err != nil {
			return err
		}

		ee, ok := e.(TestEquivocationEvidence)
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
