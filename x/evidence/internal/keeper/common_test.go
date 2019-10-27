package keeper_test

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	"gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	_   types.Evidence = (*EquivocationEvidence)(nil)
	cdc                = codec.New()
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

	EquivocationEvidence struct {
		Power      int64
		TotalPower int64
		PubKey     crypto.PubKey
		VoteA      SimpleVote
		VoteB      SimpleVote
	}
)

func init() {
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	types.RegisterEvidenceTypeCodec(EquivocationEvidence{}, "test/EquivocationEvidence")
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
		return fmt.Errorf("invalid evidence; H/R/S does not match (got %v and %v)", e.VoteA, e.VoteB)
	}

	if !bytes.Equal(e.VoteA.ValidatorAddress, e.VoteB.ValidatorAddress) {
		return fmt.Errorf(
			"invalid evidence; validator addresses do not match (got %X and %X)",
			e.VoteA.ValidatorAddress,
			e.VoteB.ValidatorAddress,
		)
	}

	if !e.PubKey.VerifyBytes(e.VoteA.SignBytes(), e.VoteA.Signature) {
		return errors.New("invalid evidence; failed to verify vote A signature")
	}
	if !e.PubKey.VerifyBytes(e.VoteB.SignBytes(), e.VoteB.Signature) {
		return errors.New("invalid evidence; failed to verify vote B signature")
	}

	return nil
}

func (e EquivocationEvidence) Hash() cmn.HexBytes {
	return tmhash.Sum(cdc.MustMarshalBinaryBare(e))
}

func (v SimpleVote) SignBytes() []byte {
	bz, _ := cdc.MarshalBinaryLengthPrefixed(v)
	return bz
}

func EquivocationHandler(k keeper.Keeper) types.Handler {
	return func(ctx sdk.Context, e types.Evidence) error {
		if err := e.ValidateBasic(); err != nil {
			return err
		}

		// TODO: Slashing!

		return nil
	}
}
