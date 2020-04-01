package types

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/version"
)

// DefaultKeyringServiceName defines a default service name for the keyring.
const DefaultKeyringServiceName = "cosmos"

// Config is the structure that holds the SDK configuration parameters.
// This could be used to initialize certain configuration parameters for the SDK.
type Config struct {
	fullFundraiserPath  string
	bech32AddressPrefix map[string]string
	txEncoder           TxEncoder
	addressVerifier     func([]byte) error
	coinType            uint32

	sync.RWMutex
}

// cosmos-sdk wide global singleton
var sdkConfig *Config

// GetConfig returns the current config instance for the SDK, or sets it with defaults
func GetConfig() *Config {
	if sdkConfig != nil {
		return sdkConfig
	}

	sdkConfig = &Config{
		bech32AddressPrefix: map[string]string{
			"account_addr":   Bech32PrefixAccAddr,
			"validator_addr": Bech32PrefixValAddr,
			"consensus_addr": Bech32PrefixConsAddr,
			"account_pub":    Bech32PrefixAccPub,
			"validator_pub":  Bech32PrefixValPub,
			"consensus_pub":  Bech32PrefixConsPub,
		},
		coinType:           CoinType,
		fullFundraiserPath: FullFundraiserPath,
		txEncoder:          nil,
	}

	return sdkConfig
}

// SetBech32PrefixForAccount builds the Config with Bech32 addressPrefix and publKeyPrefix for accounts
// and returns the config instance
func (config *Config) SetBech32PrefixForAccount(addressPrefix, pubKeyPrefix string) {
	config.Lock()
	defer config.Unlock()
	config.bech32AddressPrefix["account_addr"] = addressPrefix
	config.bech32AddressPrefix["account_pub"] = pubKeyPrefix
}

// SetBech32PrefixForValidator builds the Config with Bech32 addressPrefix and publKeyPrefix for validators
//  and returns the config instance
func (config *Config) SetBech32PrefixForValidator(addressPrefix, pubKeyPrefix string) {
	config.Lock()
	defer config.Unlock()
	config.bech32AddressPrefix["validator_addr"] = addressPrefix
	config.bech32AddressPrefix["validator_pub"] = pubKeyPrefix
}

// SetBech32PrefixForConsensusNode builds the Config with Bech32 addressPrefix and publKeyPrefix for consensus nodes
// and returns the config instance
func (config *Config) SetBech32PrefixForConsensusNode(addressPrefix, pubKeyPrefix string) {
	config.Lock()
	defer config.Unlock()
	config.bech32AddressPrefix["consensus_addr"] = addressPrefix
	config.bech32AddressPrefix["consensus_pub"] = pubKeyPrefix
}

// SetTxEncoder builds the Config with TxEncoder used to marshal StdTx to bytes
func (config *Config) SetTxEncoder(encoder TxEncoder) {
	config.Lock()
	defer config.Unlock()
	config.txEncoder = encoder
}

// SetAddressVerifier builds the Config with the provided function for verifying that addresses
// have the correct format
func (config *Config) SetAddressVerifier(addressVerifier func([]byte) error) {
	config.Lock()
	defer config.Unlock()
	config.addressVerifier = addressVerifier
}

// SetCoinType the BIP-0044 CoinType code on the config
func (config *Config) SetCoinType(coinType uint32) {
	config.Lock()
	defer config.Unlock()
	config.coinType = coinType
}

// SetFullFundraiserPath the FullFundraiserPath (BIP44Prefix) on the config
func (config *Config) SetFullFundraiserPath(fullFundraiserPath string) {
	config.Lock()
	defer config.Unlock()
	config.fullFundraiserPath = fullFundraiserPath
}

// GetBech32AccountAddrPrefix returns the Bech32 prefix for account address
func (config *Config) GetBech32AccountAddrPrefix() (out string) {
	config.RLock()
	out = config.bech32AddressPrefix["account_addr"]
	config.RUnlock()
	return
}

// GetBech32ValidatorAddrPrefix returns the Bech32 prefix for validator address
func (config *Config) GetBech32ValidatorAddrPrefix() (out string) {
	config.RLock()
	out = config.bech32AddressPrefix["validator_addr"]
	config.RUnlock()
	return
}

// GetBech32ConsensusAddrPrefix returns the Bech32 prefix for consensus node address
func (config *Config) GetBech32ConsensusAddrPrefix() (out string) {
	config.RLock()
	out = config.bech32AddressPrefix["consensus_addr"]
	config.RUnlock()
	return
}

// GetBech32AccountPubPrefix returns the Bech32 prefix for account public key
func (config *Config) GetBech32AccountPubPrefix() (out string) {
	config.RLock()
	out = config.bech32AddressPrefix["account_pub"]
	config.RUnlock()
	return
}

// GetBech32ValidatorPubPrefix returns the Bech32 prefix for validator public key
func (config *Config) GetBech32ValidatorPubPrefix() (out string) {
	config.RLock()
	out = config.bech32AddressPrefix["validator_pub"]
	config.RUnlock()
	return
}

// GetBech32ConsensusPubPrefix returns the Bech32 prefix for consensus node public key
func (config *Config) GetBech32ConsensusPubPrefix() (out string) {
	config.RLock()
	out = config.bech32AddressPrefix["consensus_pub"]
	config.RUnlock()
	return
}

// GetTxEncoder return function to encode transactions
func (config *Config) GetTxEncoder() (out TxEncoder) {
	config.RLock()
	out = config.txEncoder
	config.RUnlock()
	return
}

// GetAddressVerifier returns the function to verify that addresses have the correct format
func (config *Config) GetAddressVerifier() (out func([]byte) error) {
	config.RLock()
	out = config.addressVerifier
	config.RUnlock()
	return
}

// GetCoinType returns the BIP-0044 CoinType code on the config.
func (config *Config) GetCoinType() (out uint32) {
	config.RLock()
	out = config.coinType
	config.RUnlock()
	return
}

// GetFullFundraiserPath returns the BIP44Prefix.
func (config *Config) GetFullFundraiserPath() (out string) {
	config.RLock()
	out = config.fullFundraiserPath
	config.RUnlock()
	return
}

// KeyringServiceName returns the passed in keyring service or the default
func KeyringServiceName() string {
	if len(version.Name) == 0 {
		return DefaultKeyringServiceName
	}
	return version.Name
}
