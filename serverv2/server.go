package serverv2

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Service interface {
	Start() error
	Stop() error
}

type HasCLICommands interface {
	CLICommands() []*cobra.Command
}

type HasConfig interface {
	Config() *viper.Viper
}

var _ Service = (*Server)(nil)

type Server struct {
	services []Service
}

func NewServer(services ...Service) *Server {
	return &Server{
		services: services,
	}
}

func (s *Server) Start() error {
	for _, service := range s.services {
		go service.Start()
	}

	return nil
}

func (s *Server) Stop() error {
	for _, service := range s.services {
		if err := service.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) CLICommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, service := range s.services {
		if cliService, ok := service.(HasCLICommands); ok {
			commands = append(commands, cliService.CLICommands()...)
		}
	}

	return commands
}

func (s *Server) Configs() (*viper.Viper, error) {
	v := viper.New()

	var err error
	for _, service := range s.services {
		if configService, ok := service.(HasConfig); ok {
			err = errors.Join(err, v.MergeConfigMap(configService.Config().AllSettings()))
		}
	}
	if err != nil {
		return nil, err
	}

	return v, nil
}
