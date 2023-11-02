package serverv2

import (
	"cosmossdk.io/log"

	"github.com/spf13/cobra"
)

type Service interface {
	Start() error
	Stop() error

	CLICommands() []*cobra.Command
}

var _ Service = (*Server)(nil)

type Server struct {
	logger   log.Logger
	services []Service
}

func NewServer(logger log.Logger, services ...Service) *Server {
	return &Server{
		logger:   logger,
		services: services,
	}
}

func (s *Server) Start() error {
	for _, service := range s.services {
		if err := service.Start(); err != nil {
			return err
		}
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
		commands = append(commands, service.CLICommands()...)
	}

	return commands
}
