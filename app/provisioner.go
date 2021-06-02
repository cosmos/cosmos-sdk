package app

import "github.com/cosmos/cosmos-sdk/container"

type Provisioner interface {
	Provision(registrar container.Registrar) error
}
