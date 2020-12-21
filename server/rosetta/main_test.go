// +build rosetta_integration_tests

package rosetta

import (
	"fmt"
	"os"
	"testing"
)

// integrationClient is the client spawned in an integration environment
// which can connect to the integration cosmos gRPC and tendermint RPC
var integrationClient *Client

func TestMain(m *testing.M) {
	err := makeTestPreconditions()
	if err != nil {
		fmt.Println("precondition failed:", err.Error())
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func makeTestPreconditions() error {
	var err error
	integrationClient, err = NewOnlineServicer("localhost:9090", "tcp://localhost:26657")
	if err != nil {
		return fmt.Errorf("client init failure: %w", err)
	}
	return nil
}
