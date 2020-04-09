package keys

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCommands(t *testing.T) {
	rootCommands := Commands()
	assert.NotNil(t, rootCommands)

	// Commands are registered
	assert.Equal(t, 10, len(rootCommands.Commands()))
}

func TestMain(m *testing.M) {
	viper.Set(flags.FlagKeyringBackend, keyring.BackendTest)
	viper.Set(flagCoinType, sdk.CoinType)
	os.Exit(m.Run())
}
