package accounts

import (
	"testing"

	"cosmossdk.io/simapp"
)

func setupApp(t *testing.T) *simapp.SimApp {
	app := simapp.Setup(t, false)
	return app
}
