package client

import (
	"cosmossdk.io/server/v2/core/appmanager"
	txsigning "cosmossdk.io/x/tx/signing"
)

type ClientContext interface {
	// InterfaceRegistry returns the InterfaceRegistry.
	InterfaceRegistry() appmanager.InterfaceRegistry
	ChainID() string
	TxConfig() TxConfig
}

type TxConfig interface {
	SignModeHandler() *txsigning.HandlerMap
}
