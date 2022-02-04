package config_test

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/app/config"
	// Load the app module implementation:
	_ "github.com/cosmos/cosmos-sdk/app/internal/moduleimpl"
	"github.com/cosmos/cosmos-sdk/container"
)

func TestConfig(t *testing.T) {
	bz, err := os.ReadFile("testdata/config1.yaml")
	assert.NilError(t, err)
	containerOpt := config.LoadYAML(bz)
	err = container.Run(func() {}, containerOpt)
	assert.NilError(t, err)
}
