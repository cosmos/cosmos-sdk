package internal

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/app"

	"github.com/cosmos/cosmos-sdk/container"

	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type AppProvider struct {
	config            *app.Config
	container         *container.Container
	interfaceRegistry codecTypes.InterfaceRegistry
	codec             codec.Codec
	txConfig          client.TxConfig
	amino             *codec.LegacyAmino
	configMap         map[string]interface{}
	moduleConfigMap   map[string]*app.ModuleConfig
}

func (ap *AppProvider) InterfaceRegistry() codecTypes.InterfaceRegistry {
	return ap.interfaceRegistry
}

func (ap *AppProvider) Codec() codec.Codec {
	return ap.codec
}

func (ap *AppProvider) TxConfig() client.TxConfig {
	return ap.txConfig
}

func (ap *AppProvider) Amino() *codec.LegacyAmino {
	return ap.amino
}

func NewApp(config *app.Config) (*AppProvider, error) {
	// create container
	ctr := container.NewContainer()

	// provide encoding configs
	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)
	amino := codec.NewLegacyAmino()
	err := ctr.Provide(func() (codecTypes.InterfaceRegistry, codec.Codec, client.TxConfig, *codec.LegacyAmino) {
		return interfaceRegistry, marshaler, txConfig, amino
	})
	if err != nil {
		return nil, err
	}

	// load configs
	cfgMap := map[string]interface{}{}
	moduleConfigMap := map[string]*app.ModuleConfig{}

	for _, modConfig := range config.Modules {
		// unpack Any
		msgTyp := proto.MessageType(modConfig.Config.TypeUrl)
		cfg := reflect.New(msgTyp).Interface().(proto.Message)
		if err := proto.Unmarshal(modConfig.Config.Value, cfg); err != nil {
			return nil, err
		}

		// resolve module name
		name := modConfig.Name
		if name == "" {
			if named, ok := cfg.(Named); ok {
				name = named.Name()
			} else {
				return nil, fmt.Errorf("unnamed module config %+v", modConfig)
			}
		}

		// save in config map
		cfgMap[name] = cfg
		moduleConfigMap[name] = modConfig

		// register types
		if typeProvider, ok := cfg.(app.TypeProvider); ok {
			typeProvider.RegisterTypes(interfaceRegistry)
		}

		// register DI providers
		if provisioner, ok := cfg.(app.Provisioner); ok {
			err = provisioner.Provision(ctr)
			if err != nil {
				return nil, err
			}
		}
	}

	return &AppProvider{
		config:            config,
		container:         ctr,
		interfaceRegistry: interfaceRegistry,
		codec:             marshaler,
		txConfig:          txConfig,
		amino:             amino,
		configMap:         cfgMap,
		moduleConfigMap:   moduleConfigMap,
	}, nil
}
