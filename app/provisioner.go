package app

import "github.com/cosmos/cosmos-sdk/container"

type Provisioner interface {
	Provision(key ModuleKey) container.Option
}

type Provider interface {
	Provide(key ModuleKey) ([]interface{}, error)
}
