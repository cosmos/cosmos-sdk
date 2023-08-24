package internal

import (
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/tools/hubl/internal/config"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
)

// getAddressCodecFromConfig returns the address codecs for the given chain name
func getAddressCodecFromConfig(cfg *config.Config, chainName string) (address.Codec, address.Codec, address.Codec, error) {
	addressPrefix := "cosmos"

	if chainName != config.GlobalKeyringDirName {
		chainConfig, ok := cfg.Chains[chainName]
		if !ok {
			return nil, nil, nil, fmt.Errorf("chain %s not found in config", chainName)
		}

		addressPrefix = chainConfig.AddressPrefix
	}

	return addresscodec.NewBech32Codec(addressPrefix),
		addresscodec.NewBech32Codec(fmt.Sprintf("%svaloper", addressPrefix)),
		addresscodec.NewBech32Codec(fmt.Sprintf("%svalcons", addressPrefix)),
		nil
}
