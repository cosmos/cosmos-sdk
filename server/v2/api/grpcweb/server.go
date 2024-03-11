package grpcweb

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"cosmossdk.io/log"
)

type Server struct {
	logger  log.Logger
	GRPCSrv *grpc.Server
	unsafe  bool
	config  Config
}

func New(logger log.Logger, grpcSrv *grpc.Server, cfg Config, unsafe bool) *Server {
	return &Server{
		logger:  logger,
		GRPCSrv: grpcSrv,
		unsafe:  unsafe,
		config:  cfg,
	}
}

// Register implements registers a grpc-web server
func (s *Server) Register(r mux.Router) error {
	// configure grpc-web server
	if s.config.Enable {
		var options []grpcweb.Option
		if s.unsafe {
			options = append(options,
				grpcweb.WithOriginFunc(func(origin string) bool {
					return true
				}),
			)
		}

		wrappedGrpc := grpcweb.WrapServer(s.GRPCSrv, options...)
		r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(w, req)
				return
			}
		}))
	}
	return nil
}
