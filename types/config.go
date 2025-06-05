package types

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/cosmos/cosmos-sdk/version"
)

const (
	DefaultKeyringServiceName = "cosmos"
	EnvConfigScope            = "COSMOS_SDK_CONFIG_SCOPE"
)

type Config struct {
	fullFundraiserPath  string
	bech32AddressPrefix map[string]string
	txEncoder           TxEncoder
	addressVerifier     func([]byte) error
	mtx                 sync.RWMutex

	purpose  uint32
	coinType uint32

	sealed   bool
	sealedch chan struct{}
}

var (
	configRegistry = make(map[string]*Config)
	registryMutex  sync.Mutex
)

// getConfigKey returns a unique config scope identifier.
// It uses ENV override, or defaults to "hostname|binary|pid".
func getConfigKey() string {
	if id := os.Getenv(EnvConfigScope); id != "" {
		return id
	}

	exe, errExec := os.Executable()
	host, errHost := os.Hostname()
	pid := os.Getpid()

	if errExec != nil {
		exe = "unknown-exe"
	}
	if errHost != nil {
		host = "unknown-host"
	}

	return fmt.Sprintf("%s|%s|%d", host, exe, pid)
}

// NewConfig returns a new Config with default values.
func NewConfig() *Config {
	return &Config{
		sealedch: make(chan struct{}),
		bech32AddressPrefix: map[string]string{
			"account_addr":   Bech32PrefixAccAddr,
			"validator_addr": Bech32PrefixValAddr,
			"consensus_addr": Bech32PrefixConsAddr,
			"account_pub":    Bech32PrefixAccPub,
			"validator_pub":  Bech32PrefixValPub,
			"consensus_pub":  Bech32PrefixConsPub,
		},
		fullFundraiserPath: FullFundraiserPath,
		purpose:            Purpose,
		coinType:           CoinType,
	}
}

// GetConfig returns a per-scope config instance.
func GetConfig() *Config {
	key := getConfigKey()

	registryMutex.Lock()
	defer registryMutex.Unlock()

	if cfg, exists := configRegistry[key]; exists {
		return cfg
	}

	cfg := NewConfig()
	configRegistry[key] = cfg

	return cfg
}

func (config *Config) assertNotSealed() {
	config.mtx.RLock()
	defer config.mtx.RUnlock()
	if config.sealed {
		panic("Config is sealed")
	}
}

func (config *Config) SetBech32PrefixForAccount(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["account_addr"] = addressPrefix
	config.bech32AddressPrefix["account_pub"] = pubKeyPrefix
}

func (config *Config) SetBech32PrefixForValidator(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["validator_addr"] = addressPrefix
	config.bech32AddressPrefix["validator_pub"] = pubKeyPrefix
}

func (config *Config) SetBech32PrefixForConsensusNode(addressPrefix, pubKeyPrefix string) {
	config.assertNotSealed()
	config.bech32AddressPrefix["consensus_addr"] = addressPrefix
	config.bech32AddressPrefix["consensus_pub"] = pubKeyPrefix
}

func (config *Config) SetTxEncoder(encoder TxEncoder) {
	config.assertNotSealed()
	config.txEncoder = encoder
}

func (config *Config) SetAddressVerifier(addressVerifier func([]byte) error) {
	config.assertNotSealed()
	config.addressVerifier = addressVerifier
}

// SetFullFundraiserPath sets the FullFundraiserPath (BIP44Prefix) on the config.
//
// Deprecated: This method is supported for backward compatibility only and will be removed in a future release. Use SetPurpose and SetCoinType instead.
func (config *Config) SetFullFundraiserPath(fullFundraiserPath string) {
	config.assertNotSealed()
	config.fullFundraiserPath = fullFundraiserPath
}

// SetPurpose sets the BIP-0044 Purpose code on the config
func (config *Config) SetPurpose(purpose uint32) {
	config.assertNotSealed()
	config.purpose = purpose
}

// SetCoinType sets the BIP-0044 CoinType code on the config
func (config *Config) SetCoinType(coinType uint32) {
	config.assertNotSealed()
	config.coinType = coinType
}

func (config *Config) Seal() *Config {
	config.mtx.Lock()
	defer config.mtx.Unlock()

	if config.sealed {
		return config
	}

	config.sealed = true
	close(config.sealedch)

	return config
}

func (config *Config) GetBech32AccountAddrPrefix() string {
	return config.bech32AddressPrefix["account_addr"]
}

func (config *Config) GetBech32ValidatorAddrPrefix() string {
	return config.bech32AddressPrefix["validator_addr"]
}

func (config *Config) GetBech32ConsensusAddrPrefix() string {
	return config.bech32AddressPrefix["consensus_addr"]
}

func (config *Config) GetBech32AccountPubPrefix() string {
	return config.bech32AddressPrefix["account_pub"]
}

func (config *Config) GetBech32ValidatorPubPrefix() string {
	return config.bech32AddressPrefix["validator_pub"]
}

func (config *Config) GetBech32ConsensusPubPrefix() string {
	return config.bech32AddressPrefix["consensus_pub"]
}

func (config *Config) GetTxEncoder() TxEncoder {
	return config.txEncoder
}

func (config *Config) GetAddressVerifier() func([]byte) error {
	return config.addressVerifier
}

func (config *Config) GetPurpose() uint32 {
	return config.purpose
}

func (config *Config) GetCoinType() uint32 {
	return config.coinType
}

func (config *Config) GetFullFundraiserPath() string {
	return config.fullFundraiserPath
}

func (config *Config) GetFullBIP44Path() string {
	return fmt.Sprintf("m/%d'/%d'/0'/0/0", config.purpose, config.coinType)
}

func KeyringServiceName() string {
	if len(version.Name) == 0 {
		return DefaultKeyringServiceName
	}
	return version.Name
}

// Optional: expose sealed config with timeout

func GetSealedConfig(ctx context.Context) (*Config, error) {
	config := GetConfig()
	select {
	case <-config.sealedch:
		return config, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
