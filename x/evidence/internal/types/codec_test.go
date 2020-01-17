package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
)

var _ exported.Evidence = (*testEvidence)(nil)

type testEvidence struct{}

func (te testEvidence) Route() string                        { return "" }
func (te testEvidence) Type() string                         { return "" }
func (te testEvidence) String() string                       { return "" }
func (te testEvidence) ValidateBasic() error                 { return nil }
func (te testEvidence) GetConsensusAddress() sdk.ConsAddress { return nil }
func (te testEvidence) Hash() tmbytes.HexBytes               { return nil }
func (te testEvidence) GetHeight() int64                     { return 0 }
func (te testEvidence) GetValidatorPower() int64             { return 0 }
func (te testEvidence) GetTotalPower() int64                 { return 0 }

func TestCodec(t *testing.T) {
	cdc := codec.New()
	types.RegisterCodec(cdc)
	types.RegisterEvidenceTypeCodec(testEvidence{}, "cosmos-sdk/testEvidence")

	var e exported.Evidence = testEvidence{}
	bz, err := cdc.MarshalBinaryBare(e)
	require.NoError(t, err)

	var te testEvidence
	require.NoError(t, cdc.UnmarshalBinaryBare(bz, &te))

	require.Panics(t, func() { types.RegisterEvidenceTypeCodec(testEvidence{}, "cosmos-sdk/testEvidence") })
}
