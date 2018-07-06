package types

import (
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"os"
)

// server context
type ServerContext struct {
	Config *cfg.Config
	Logger log.Logger
}

func NewDefaultServerContext() *ServerContext {
	return NewServerContext(
		cfg.DefaultConfig(),
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	)
}

func NewServerContext(config *cfg.Config, logger log.Logger) *ServerContext {
	return &ServerContext{config, logger}
}
