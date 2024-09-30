package rest

import (
	"context"

	"github.com/gorilla/mux"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

const (
	ServerName = "rest-v2"
)

type Server[T transaction.Tx] struct {
	logger log.Logger
	router *mux.Router
}

func New[T transaction.Tx]() *Server[T] {
	return &Server[T]{
		router: mux.NewRouter(),
	}
}

func (s *Server[T]) Name() string {
	return ServerName
}

func (s *Server[T]) Start(ctx context.Context) error {
	return nil
}

func (s *Server[T]) Stop(ctx context.Context) error {
	return nil
}

func (s *Server[T]) Init(appI serverv2.AppI[T], cfg map[string]any, logger log.Logger) error {
	s.logger = logger.With(log.ModuleKey, s.Name())

	return nil
}
