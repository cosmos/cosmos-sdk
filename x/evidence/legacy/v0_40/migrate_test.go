package v040_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
	v040evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_40"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func TestMigrate(t *testing.T) {
	v040Codec := codec.New()
	cryptocodec.RegisterCrypto(v040Codec)

	addr1, _ := sdk.AccAddressFromBech32("cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u")

	evidenceGenState := v038evidence.GenesisState{
		Params: v038evidence.Params{MaxEvidenceAge: v038evidence.DefaultMaxEvidenceAge},
		Evidence: []exported.Evidence{&types.Equivocation{
			Height:           20,
			Power:            100,
			Time:             time.Date(2020, 01, 01, 01, 01, 01, 01, time.Local).UTC(),
			ConsensusAddress: addr1.Bytes(),
		}},
	}

	migrated := v040evidence.Migrate(evidenceGenState)
	expected := `{
  "evidence": [
    {
      "height": "20",
      "time": "2020-01-01T00:01:01.000000001Z",
      "power": "100",
      "consensus_address": "cosmosvalcons1xxkueklal9vejv9unqu80w9vptyepfa99x2a3w"
    }
  ]
}`

	bz, err := v040Codec.MarshalJSONIndent(migrated, "", "  ")
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
