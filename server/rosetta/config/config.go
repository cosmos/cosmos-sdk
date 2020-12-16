package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/spf13/pflag"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/client"
	"github.com/cosmos/cosmos-sdk/server/rosetta/services"
)

// configuration defaults constants
const (
	// DefaultBlockchain defines the default blockchain identifier name
	DefaultBlockchain = "app"
	// DefaultAddr defines the default rosetta binding address
	DefaultAddr = ":8080"
	// DefaultRetries is the default number of retries
	DefaultRetries = 5
	// DefaultTendermintEndpoint is the default value for the tendermint endpoint
	DefaultTendermintEndpoint = "localhost:26657"
	// DefaultGRPCEndpoint is the default value for the gRPC endpoint
	DefaultGRPCEndpoint = "localhost:9090"
	// DefaultNetwork defines the default network name
	DefaultNetwork = "network"
	// DefaultOffline defines the default offline value
	DefaultOffline = false
)

// configuration flags
const (
	FlagBlockchain         = "blockchain"
	FlagNetwork            = "network"
	FlagTendermintEndpoint = "tendermint"
	FlagGRPCEndpoint       = "grpc"
	FlagAddr               = "addr"
	FlagRetries            = "retries"
	FlagOffline            = "offline"
)

// RosettaFromConfig builds the rosetta servicer full implementation from configurations
func RosettaFromConfig(conf *Config) (crg.Adapter, rosetta.NodeClient, error) {
	if conf.Offline {
		return services.NewOffline(conf.NetworkIdentifier()), nil, nil
	}
	var dataAPIOpts []client.OptionFunc
	if conf.codec != nil && conf.ir != nil {
		dataAPIOpts = append(dataAPIOpts, client.WithCodec(conf.ir, conf.codec))
	}
	dataAPIClient, err := client.NewSingle(conf.GRPCEndpoint, conf.TendermintRPC, dataAPIOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("data api client init failure: %w", err)
	}
	sn, err := services.NewSingleNetwork(dataAPIClient, &types.NetworkIdentifier{
		Blockchain: conf.Blockchain,
		Network:    conf.Network,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("rosetta network initialization failure: %w", err)
	}
	return sn, dataAPIClient, nil
}

// RetryRosettaFromConfig tries to initialize rosetta retrying
func RetryRosettaFromConfig(conf *Config) (rosetta crg.Adapter, client rosetta.NodeClient, err error) {
	for i := 0; i < conf.Retries; i++ {
		rosetta, client, err = RosettaFromConfig(conf)
		if err == nil {
			return
		}
		time.Sleep(5 * time.Second)
	}
	return
}

// Config defines the configuration of the rosetta server
type Config struct {
	// Blockchain defines the blockchain name
	// defaults to DefaultBlockchain
	Blockchain string
	// Network defines the network name
	Network string
	// TendermintRPC defines the endpoint to connect to
	// tendermint RPC, specifying 'tcp://' before is not
	// required, usually it's at port 26657 of the
	TendermintRPC string
	// GRPCEndpoint defines the cosmos application gRPC endpoint
	// usually it is located at 9090 port
	GRPCEndpoint string
	// Addr defines the default address to bind the rosetta server to
	// defaults to DefaultAddr
	Addr string
	// Retries defines the maximum number of retries
	// rosetta will do before quitting
	Retries int
	// Offline defines if the server must be run in offline mode
	Offline bool
	// codec overrides the default data and construction api client codecs
	codec *codec.ProtoCodec
	// ir overrides the default data and construction api interface registry
	ir codectypes.InterfaceRegistry
}

// NetworkIdentifier returns the network identifier given the configuration
func (c *Config) NetworkIdentifier() *types.NetworkIdentifier {
	return &types.NetworkIdentifier{
		Blockchain: c.Blockchain,
		Network:    c.Network,
	}
}

// Validate validates a configuration and sets
// its defaults in case they were not provided
func (c *Config) Validate() error {
	if (c.codec == nil) != (c.ir == nil) {
		return fmt.Errorf("codec and interface registry must be both different from nil or nil")
	}

	if c.Addr == "" {
		c.Addr = DefaultAddr
	}
	if c.Blockchain == "" {
		c.Blockchain = DefaultBlockchain
	}
	if c.Retries == 0 {
		c.Retries = DefaultRetries
	}
	// these are must
	if c.Network == "" {
		return fmt.Errorf("network not provided")
	}
	if c.Offline {
		return nil
	}

	// these are optional but it must be online
	if c.GRPCEndpoint == "" {
		return fmt.Errorf("grpc endpoint not provided")
	}
	if c.TendermintRPC == "" {
		return fmt.Errorf("tendermint rpc not provided")
	}
	if !strings.HasPrefix(c.TendermintRPC, "tcp://") {
		c.TendermintRPC = fmt.Sprintf("tcp://%s", c.TendermintRPC)
	}

	return nil
}

// WithCodec extends the configuration with a predefined codec
func (c *Config) WithCodec(ir codectypes.InterfaceRegistry, cdc *codec.ProtoCodec) {
	c.codec = cdc
	c.ir = ir
}

// FromFlags gets the configuration from flags
func FromFlags(flags *pflag.FlagSet) (*Config, error) {
	blockchain, err := flags.GetString(FlagBlockchain)
	if err != nil {
		return nil, err
	}
	network, err := flags.GetString(FlagNetwork)
	if err != nil {
		return nil, err
	}
	tendermintRPC, err := flags.GetString(FlagTendermintEndpoint)
	if err != nil {
		return nil, err
	}
	gRPCEndpoint, err := flags.GetString(FlagGRPCEndpoint)
	if err != nil {
		return nil, err
	}
	addr, err := flags.GetString(FlagAddr)
	if err != nil {
		return nil, err
	}
	retries, err := flags.GetInt(FlagRetries)
	if err != nil {
		return nil, err
	}
	offline, err := flags.GetBool(FlagOffline)
	if err != nil {
		return nil, err
	}
	conf := &Config{
		Blockchain:    blockchain,
		Network:       network,
		TendermintRPC: tendermintRPC,
		GRPCEndpoint:  gRPCEndpoint,
		Addr:          addr,
		Retries:       retries,
		Offline:       offline,
	}
	err = conf.Validate()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// SetFlags sets the configuration flags to the given flagset
func SetFlags(flags *pflag.FlagSet) {
	flags.String(FlagBlockchain, DefaultBlockchain, "the blockchain type")
	flags.String(FlagNetwork, DefaultNetwork, "the network name")
	flags.String(FlagTendermintEndpoint, DefaultTendermintEndpoint, "the tendermint rpc endpoint, without tcp://")
	flags.String(FlagGRPCEndpoint, DefaultGRPCEndpoint, "the app gRPC endpoint")
	flags.String(FlagAddr, DefaultAddr, "the address rosetta will bind to")
	flags.Int(FlagRetries, DefaultRetries, "the number of retries that will be done before quitting")
	flags.Bool(FlagOffline, DefaultOffline, "run rosetta only with construction API")
}
