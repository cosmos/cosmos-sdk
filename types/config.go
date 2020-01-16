package types

import (
	"sync"
)

// DefaultKeyringServiceName defines a default service name for the keyring.
const DefaultKeyringServiceName = "cosmos"

// Config is the structure that holds the SDK configuration parameters.
// This could be used to initialize certain configuration parameters for the SDK.
type Config struct {
	fullFundraiserPath  string
	keyringServiceName  string
	bech32AddressPrefix map[string]string
	txEncoder           TxEncoder
	addressVerifier     func([]byte) error
	mtx                 sync.RWMutex
	coinType            uint32
	sealed              bool
}

// NewConfig returns a new unsealed Config.
func NewConfig(fullFundraiserPath, keyringServiceName string, bech32AddressPrefix map[string]string, coinType uint32) *Config {
	return &Config{
		fullFundraiserPath:  fullFundraiserPath,
		keyringServiceName:  keyringServiceName,
		bech32AddressPrefix: bech32AddressPrefix,
		coinType:            coinType,
		sealed:              false,
		txEncoder:           nil,
	}
}

// NewDefaultConfig returns a new Config with default values.
func NewDefaultConfig() *Config {
	addressPrefixMap := NewBech32PrefixMap(
		Bech32PrefixAccAddr, Bech32PrefixValAddr, Bech32PrefixConsAddr,
		Bech32PrefixAccPub, Bech32PrefixValPub, Bech32PrefixConsPub)
	return NewConfig(FullFundraiserPath, DefaultKeyringServiceName, addressPrefixMap, CoinType)
}

// cosmos-sdk wide global singleton
var sdkConfig *Config

// GetConfig returns the config instance for the SDK.
func GetConfig() *Config {
	if sdkConfig != nil {
		return sdkConfig
	}
	sdkConfig = NewDefaultConfig()
	return sdkConfig
}

// NewBech32PrefixMap returns a new prefix map to be used for addresses bech32 representations.
func NewBech32PrefixMap(accAddr, valAddr, consAddr, accPub, valPub, consPub string) map[string]string {
	return map[string]string{
		"account_addr":   accAddr,
		"validator_addr": valAddr,
		"consensus_addr": consAddr,
		"account_pub":    accPub,
		"validator_pub":  valPub,
		"consensus_pub":  consPub,
	}
}

// NewDefaultBech32PrefixMap returns the SDK's default prefix map for addresses bech32
// representations.
func NewDefaultBech32PrefixMap() map[string]string {
	return NewBech32PrefixMap(
		Bech32PrefixAccAddr, Bech32PrefixValAddr,
		Bech32PrefixConsAddr, Bech32PrefixAccPub,
		Bech32PrefixValPub, Bech32PrefixConsPub,
	)
}

func (config *Config) assertNotSealed() {
	config.mtx.Lock()
	defer config.mtx.Unlock()

	if config.sealed {
		panic("Config is sealed")
	}
}

// SetBech32PrefixForAccount builds the Config with Bech32 addressPrefix and publKeyPrefix for accounts
// and returns the config instance
func (config *Config) SetBech32PrefixForAccount(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["account_addr"] = addressPrefix
	config.bech32AddressPrefix["account_pub"] = pubKeyPrefix
}

// SetBech32PrefixForValidator builds the Config with Bech32 addressPrefix and publKeyPrefix for validators
//  and returns the config instance
func (config *Config) SetBech32PrefixForValidator(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["validator_addr"] = addressPrefix
	config.bech32AddressPrefix["validator_pub"] = pubKeyPrefix
}

// SetBech32PrefixForConsensusNode builds the Config with Bech32 addressPrefix and publKeyPrefix for consensus nodes
// and returns the config instance
func (config *Config) SetBech32PrefixForConsensusNode(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["consensus_addr"] = addressPrefix
	config.bech32AddressPrefix["consensus_pub"] = pubKeyPrefix
}

// SetTxEncoder builds the Config with TxEncoder used to marshal StdTx to bytes
func (config *Config) SetTxEncoder(encoder TxEncoder) {
	config.assertNotSealed()
	config.txEncoder = encoder
}

// SetAddressVerifier builds the Config with the provided function for verifying that addresses
// have the correct format
func (config *Config) SetAddressVerifier(addressVerifier func([]byte) error) {
	config.assertNotSealed()
	config.addressVerifier = addressVerifier
}

// Set the BIP-0044 CoinType code on the config
func (config *Config) SetCoinType(coinType uint32) {
	config.assertNotSealed()
	config.coinType = coinType
}

// Set the FullFundraiserPath (BIP44Prefix) on the config
func (config *Config) SetFullFundraiserPath(fullFundraiserPath string) {
	config.assertNotSealed()
	config.fullFundraiserPath = fullFundraiserPath
}

// Set the keyringServiceName (BIP44Prefix) on the config
func (config *Config) SetKeyringServiceName(keyringServiceName string) {
	config.assertNotSealed()
	config.keyringServiceName = keyringServiceName
}

// Seal seals the config such that the config state could not be modified further
func (config *Config) Seal() *Config {
	config.mtx.Lock()
	defer config.mtx.Unlock()

	config.sealed = true
	return config
}

// GetBech32AccountAddrPrefix returns the Bech32 prefix for account address
func (config *Config) GetBech32AccountAddrPrefix() string {
	return config.bech32AddressPrefix["account_addr"]
}

// GetBech32ValidatorAddrPrefix returns the Bech32 prefix for validator address
func (config *Config) GetBech32ValidatorAddrPrefix() string {
	return config.bech32AddressPrefix["validator_addr"]
}

// GetBech32ConsensusAddrPrefix returns the Bech32 prefix for consensus node address
func (config *Config) GetBech32ConsensusAddrPrefix() string {
	return config.bech32AddressPrefix["consensus_addr"]
}

// GetBech32AccountPubPrefix returns the Bech32 prefix for account public key
func (config *Config) GetBech32AccountPubPrefix() string {
	return config.bech32AddressPrefix["account_pub"]
}

// GetBech32ValidatorPubPrefix returns the Bech32 prefix for validator public key
func (config *Config) GetBech32ValidatorPubPrefix() string {
	return config.bech32AddressPrefix["validator_pub"]
}

// GetBech32ConsensusPubPrefix returns the Bech32 prefix for consensus node public key
func (config *Config) GetBech32ConsensusPubPrefix() string {
	return config.bech32AddressPrefix["consensus_pub"]
}

// GetTxEncoder return function to encode transactions
func (config *Config) GetTxEncoder() TxEncoder {
	return config.txEncoder
}

// GetAddressVerifier returns the function to verify that addresses have the correct format
func (config *Config) GetAddressVerifier() func([]byte) error {
	return config.addressVerifier
}

// GetCoinType returns the BIP-0044 CoinType code on the config.
func (config *Config) GetCoinType() uint32 {
	return config.coinType
}

// GetFullFundraiserPath returns the BIP44Prefix.
func (config *Config) GetFullFundraiserPath() string {
	return config.fullFundraiserPath
}

// GetKeyringServiceName returns the keyring service name from the config.
func (config *Config) GetKeyringServiceName() string {
	return config.keyringServiceName
}
