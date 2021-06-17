package app

import "github.com/cosmos/cosmos-sdk/types"

type ModuleKey interface {
	moduleKey()
	ID() ModuleID
}

type ModuleID interface {
	moduleID()
	Name() string
}

type KVStoreKeyProvider func(ModuleKey) *types.KVStoreKey
type TransientStoreKeyProvider func(ModuleKey) *types.TransientStoreKey

type moduleKey struct {
	*moduleID
}

type moduleID struct {
	name string
}

var _ ModuleKey = &moduleKey{}

func (m *moduleKey) moduleKey() {}

func (m *moduleKey) ID() ModuleID {
	return m.moduleID
}

func (m *moduleID) moduleID() {
}

func (m *moduleID) Name() string {
	return m.name
}
