package cli

import "github.com/cosmos/cosmos-sdk/client"

type DefaultHome string

type ClientContextOption func(client.Context) client.Context

func (ClientContextOption) IsAutoGroupType() {}

type AppConfig struct {
	Config   interface{}
	Template string
}
