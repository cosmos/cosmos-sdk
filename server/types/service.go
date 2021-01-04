package types

import "github.com/cosmos/cosmos-sdk/server/config"

type Server interface {
	GetService(name string) Service
	RegisterServices() error

	Start() error
	Stop() error
}

type Service interface {
	Name() string
	RegisterRoutes() bool
	Start(config.ServerConfig) error
	Stop() error
}
