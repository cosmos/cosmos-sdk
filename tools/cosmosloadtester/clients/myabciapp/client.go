package myabciapp

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
)

// CosmosClientFactory creates instances of CosmosClient
type CosmosClientFactory struct {
	txConfig client.TxConfig
}

// CosmosClientFactory implements loadtest.ClientFactory
var _ loadtest.ClientFactory = (*CosmosClientFactory)(nil)

func NewCosmosClientFactory(txConfig client.TxConfig) *CosmosClientFactory {
	return &CosmosClientFactory{
		txConfig: txConfig,
	}
}

// CosmosClient is responsible for generating transactions. Only one client
// will be created per connection to the remote Tendermint RPC endpoint, and
// each client will be responsible for maintaining its own state in a
// thread-safe manner.
type CosmosClient struct {
	txConfig client.TxConfig
}

// CosmosClient implements loadtest.Client
var _ loadtest.Client = (*CosmosClient)(nil)

func (f *CosmosClientFactory) ValidateConfig(cfg loadtest.Config) error {
	// Do any checks here that you need to ensure that the load test
	// configuration is compatible with your client.
	return nil
}

func (f *CosmosClientFactory) NewClient(cfg loadtest.Config) (loadtest.Client, error) {
	return &CosmosClient{
		txConfig: f.txConfig,
	}, nil
}

// GenerateTx must return the raw bytes that make up the transaction for your
// ABCI app. The conversion to base64 will automatically be handled by the
// loadtest package, so don't worry about that. Only return an error here if you
// want to completely fail the entire load test operation.
func (c *CosmosClient) GenerateTx() ([]byte, error) {
	txBuilder := c.txConfig.NewTxBuilder()
	// TODO: add messages to this transaction.
	return c.txConfig.TxEncoder()(txBuilder.GetTx())
}
