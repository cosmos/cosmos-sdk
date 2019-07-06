package port

import (
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Manager struct {
	protocol state.Mapping
}

func NewManager(protocol state.Base) Manager {
	return Manager
}
