package main

import (
	"fmt"
)

type Chain struct {
	Name          string  `name:"name" json:"name" yaml:"name"`
	Type          string  `name:"type" json:"type" yaml:"type"`
	NumValidators int     `name:"num-validators" json:"num_validators" yaml:"numValidators"`
	Ports         Port    `name:"ports" json:"ports" yaml:"ports"`
	Upgrade       Upgrade `name:"upgrade" json:"upgrade" yaml:"upgrade"`
}

func (c *Chain) GetRPCAddr() string {
	return fmt.Sprintf("http://localhost:%d", c.Ports.Rpc)
}

func (c *Chain) GetRESTAddr() string {
	return fmt.Sprintf("http://localhost:%d", c.Ports.Rest)
}

func (c *Chain) GetGRPCAddr() string {
	return fmt.Sprintf("http://localhost:%d", c.Ports.Grpc)
}

func (c *Chain) GetFaucetAddr() interface{} {
	return fmt.Sprintf("http://localhost:%d", c.Ports.Faucet)
}

type Upgrade struct {
	Enabled  bool   `name:"eanbled" json:"enabled" yaml:"enabled"`
	Type     string `name:"type" json:"type" yaml:"type"`
	Genesis  string `name:"genesis" json:"genesis" yaml:"genesis"`
	Upgrades []struct {
		Name    string `name:"name" json:"name" yaml:"name"`
		Version string `name:"version" json:"version" yaml:"version"`
	} `name:"upgrades" json:"upgrades" yaml:"upgrades"`
}

type Port struct {
	Rest    int `name:"rest" json:"rest" yaml:"rest"`
	Rpc     int `name:"rpc" json:"rpc" yaml:"rpc"`
	Grpc    int `name:"grpc" json:"grpc" yaml:"grpc"`
	Exposer int `name:"exposer" json:"exposer" yaml:"exposer"`
	Faucet  int `name:"faucet" json:"faucet" yaml:"faucet"`
}

type Relayer struct {
	Name     string   `name:"name" json:"name" yaml:"name"`
	Type     string   `name:"type" json:"type" yaml:"type"`
	Replicas int      `name:"replicas" json:"replicas" yaml:"replicas"`
	Chains   []string `name:"chains" json:"chains" yaml:"chains"`
}

type Feature struct {
	Enabled bool   `name:"enabled" json:"enabled" yaml:"enabled"`
	Image   string `name:"image" json:"image" yaml:"image"`
	Ports   Port   `name:"ports" json:"ports" yaml:"ports"`
}

func (f *Feature) GetRPCAddr() string {
	return fmt.Sprintf("http://localhost:%d", f.Ports.Rpc)
}

func (f *Feature) GetRESTAddr() string {
	return fmt.Sprintf("http://localhost:%d", f.Ports.Rest)
}

// Config is the struct for the config.yaml setup file
// todo: move this to a more common place, outside just tests
// todo: can be moved to proto definition
type Config struct {
	Chains   []*Chain   `name:"chains" json:"chains" yaml:"chains"`
	Relayers []*Relayer `name:"relayers" json:"relayers" yaml:"relayers"`
	Explorer *Feature   `name:"explorer" json:"explorer" yaml:"explorer"`
	Registry *Feature   `name:"registry" json:"registry" yaml:"registry"`
}

// HasChainId returns true if chain id found in list of chains
func (c *Config) HasChainId(chainId string) bool {
	for _, chain := range c.Chains {
		if chain.Name == chainId {
			return true
		}
	}

	return false
}

// GetChain returns the Chain object pointer for the given chain id
func (c *Config) GetChain(chainId string) *Chain {
	for _, chain := range c.Chains {
		if chain.Name == chainId {
			return chain
		}
	}

	return nil
}
