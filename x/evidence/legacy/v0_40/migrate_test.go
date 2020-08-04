package v040_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
	v040evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_40"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func TestMigrate(t *testing.T) {
	v040Codec := codec.New()
	v040Codec.RegisterInterface((*v038evidence.Evidence)(nil), nil)
	v040Codec.RegisterConcrete(&v038evidence.Equivocation{}, "cosmos-sdk/Equivocation", nil)
	cryptocodec.RegisterCrypto(v040Codec)

	addr1, _ := sdk.AccAddressFromBech32("cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u")

	evidenceGenState := v038evidence.GenesisState{
		Params: v038evidence.Params{MaxEvidenceAge: v038evidence.DefaultMaxEvidenceAge},
		Evidence: []v038evidence.Evidence{&types.Equivocation{
			Height:           20,
			Power:            100,
			ConsensusAddress: addr1.Bytes(),
		}},
	}

	migrated := v040evidence.Migrate(evidenceGenState)
	expected := `{
  "evidence": [
    {
      "height": "20",
      "time": "0001-01-01T00:00:00Z",
      "power": "100",
      "consensus_address": "cosmosvalcons1xxkueklal9vejv9unqu80w9vptyepfa99x2a3w"
    }
  ]
}`

	bz, err := v040Codec.MarshalJSONIndent(migrated, "", "  ")
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
