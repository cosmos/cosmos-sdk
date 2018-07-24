package lcd

import (
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestProvider(t *testing.T) {

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "Voyager",
		Provider: "LCD",
	}

	// Start provider API in the background
	password := "1234567890"
	addr, _ := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	var baseURL = "http://localhost:" + port

	// Verify the Provider with local Pact Files
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        baseURL,
		PactURLs:               []string{filepath.ToSlash("Voyager-LCD.json")},
		ProviderStatesSetupURL: baseURL + "/setup",
	})
}
