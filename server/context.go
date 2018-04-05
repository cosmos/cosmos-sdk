package server

import (
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tmlibs/log"
)

type Context struct {
	Config *cfg.Config
	Logger log.Logger
}

func NewContext(config *cfg.Config, logger log.Logger) *Context {
	return &Context{config, logger}
}
