package v040_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
	v040evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_40"
)

func TestMigrate(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()

	txCfg := encodingConfig.TxConfig
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithInterfaceRegistry(encodingConfig.InterfaceRegistry)
	clientCtx = clientCtx.WithTxConfig(txCfg)
	clientCtx = clientCtx.WithCodec(encodingConfig.Amino)
	clientCtx = clientCtx.WithJSONMarshaler(encodingConfig.Marshaler)

	addr1, _ := sdk.AccAddressFromBech32("cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u")

	evidenceGenState := v038evidence.GenesisState{
		Params: v038evidence.Params{MaxEvidenceAge: v038evidence.DefaultMaxEvidenceAge},
		Evidence: []v038evidence.Evidence{&v038evidence.Equivocation{
			Height:           20,
			Power:            100,
			ConsensusAddress: addr1.Bytes(),
		}},
	}

	migrated := v040evidence.Migrate(evidenceGenState, clientCtx)
	// Using hybrid codec here, expecting Amino JSON.
	expected := `{"evidence":[{"type":"cosmos-sdk/Equivocation","value":{"height":"20","time":"0001-01-01T00:00:00Z","power":"100","consensus_address":"cosmosvalcons1xxkueklal9vejv9unqu80w9vptyepfa99x2a3w"}}]}`

	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
