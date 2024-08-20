package server

import (
	"cosmossdk.io/core/transaction"
)

// InterfaceRegistry defines the interface for resolving interfaces
type InterfaceRegistry interface {
	AnyResolver
	ListImplementations(ifaceTypeURL string) []string
	ListAllInterfaces() []string
}

// AnyResolver defines the interface for resolving interfaces
// This is used to avoid the gogoproto import in core
type AnyResolver = interface {
	Resolve(typeUrl string) (transaction.Msg, error)
}
