package serverv2

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Module interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type HasCLICommands interface {
	CLICommands() []*cobra.Command
}

type HasConfig interface {
	Config() *viper.Viper
}

var _ Module = (*Server)(nil)

type Server struct {
	modules []Module
}

func NewServer(modules ...Module) *Server {
	return &Server{
		modules: modules,
	}
}

// Start starts all modules concurrently.
func (s *Server) Start(ctx context.Context) error {
	var err error
	for _, mod := range s.modules {
		go func(mod Module) {
			err = errors.Join(err, mod.Start(ctx))
		}(mod)
	}

	return err
}

// Stop sequentially stops all modules.
func (s *Server) Stop(ctx context.Context) error {
	for _, mod := range s.modules {
		if err := mod.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) CLICommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, mod := range s.modules {
		if climod, ok := mod.(HasCLICommands); ok {
			commands = append(commands, climod.CLICommands()...)
		}
	}

	return commands
}

func (s *Server) Configs() (*viper.Viper, error) {
	v := viper.New()

	var err error
	for _, mod := range s.modules {
		if configmod, ok := mod.(HasConfig); ok {
			err = errors.Join(err, v.MergeConfigMap(configmod.Config().AllSettings()))
		}
	}
	if err != nil {
		return nil, err
	}

	return v, nil
}
