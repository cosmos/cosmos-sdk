package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewDefaultBech32PrefixMap(t *testing.T) {
	defaultPrefixMap := sdk.NewDefaultBech32PrefixMap()
	require.Equal(t, len(defaultPrefixMap), 6)
	require.Equal(t, "cosmos", defaultPrefixMap["account_addr"])
	require.Equal(t, "cosmosvaloper", defaultPrefixMap["validator_addr"])
	require.Equal(t, "cosmosvalcons", defaultPrefixMap["consensus_addr"])
	require.Equal(t, "cosmospub", defaultPrefixMap["account_pub"])
	require.Equal(t, "cosmosvaloperpub", defaultPrefixMap["validator_pub"])
	require.Equal(t, "cosmosvalconspub", defaultPrefixMap["consensus_pub"])
}

func TestNewDefaultConfig(t *testing.T) {
	defaultConfig := sdk.NewDefaultConfig()
	require.Equal(t, "cosmos", defaultConfig.GetBech32AccountAddrPrefix())
	require.Equal(t, "cosmosvaloper", defaultConfig.GetBech32ValidatorAddrPrefix())
	require.Equal(t, "cosmosvalcons", defaultConfig.GetBech32ConsensusAddrPrefix())
	require.Equal(t, "cosmospub", defaultConfig.GetBech32AccountPubPrefix())
	require.Equal(t, "cosmosvaloperpub", defaultConfig.GetBech32ValidatorPubPrefix())
	require.Equal(t, "cosmosvalconspub", defaultConfig.GetBech32ConsensusPubPrefix())
	require.Equal(t, uint32(sdk.CoinType), defaultConfig.GetCoinType())
	require.Equal(t, sdk.DefaultKeyringServiceName, defaultConfig.GetKeyringServiceName())
	require.Equal(t, sdk.FullFundraiserPath, defaultConfig.GetFullFundraiserPath())
	require.Nil(t, defaultConfig.GetAddressVerifier())
	require.Nil(t, defaultConfig.GetTxEncoder())
	var txEnc sdk.TxEncoder = nil
	defaultConfig.SetTxEncoder(txEnc)
	defaultConfig.Seal()
	require.Panics(t, func() { defaultConfig.SetTxEncoder(txEnc) })
}
