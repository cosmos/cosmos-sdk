package types

import "github.com/cosmos/cosmos-sdk/app"

type Inputs struct {
	Handlers map[string]app.Handler
}
