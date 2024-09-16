package telemetry

import (
	"context"
	"fmt"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*TelemetryServer[transaction.Tx])(nil)
	_ serverv2.HasConfig                       = (*TelemetryServer[transaction.Tx])(nil)
)

const ServerName = "telemetry"

type TelemetryServer[T transaction.Tx] struct {
	config *Config
	logger log.Logger
}

// New creates a new telemtry server.
func New[T transaction.Tx]() *TelemetryServer[T] {
	return &TelemetryServer[T]{}
}

// Name returns the server name.
func (s *TelemetryServer[T]) Name() string {
	return ServerName
}

func (s *TelemetryServer[T]) Config() any {
	if s.config == nil || s.config == (&Config{}) {
		return DefaultConfig()
	}

	return s.config
}

// Init implements serverv2.ServerComponent.
func (s *TelemetryServer[T]) Init(appI serverv2.AppI[T], cfg map[string]any, logger log.Logger) error {
	serverCfg := s.Config().(*Config)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, s.Name(), &serverCfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	s.config = serverCfg
	s.logger = logger

	return nil
}

func (s *TelemetryServer[T]) Start(context.Context) error {
	return nil
}

func (s *TelemetryServer[T]) Stop(context.Context) error {
	return nil
}
