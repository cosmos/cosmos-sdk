package nodeservice

import (
	context "context"
	"errors"

	"cosmossdk.io/core/appmodule"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/server"
)

var _ ServiceServer = queryServer{}

type queryServer struct {
	cfg server.ConfigMap
}

func NewQueryServer(cfg server.ConfigMap) ServiceServer {
	return queryServer{cfg: cfg}
}

func (s queryServer) Config(ctx context.Context, _ *ConfigRequest) (*ConfigResponse, error) {
	minGasPricesStr := ""
	minGasPrices, ok := s.cfg["server.minimum_gas_price"]
	if !ok {
		minGasPricesStr = minGasPrices.(string)
	}

	return &ConfigResponse{
		MinimumGasPrice: minGasPricesStr,
	}, nil
}

func (s queryServer) Status(ctx context.Context, _ *StatusRequest) (*StatusResponse, error) {
	environment, ok := ctx.Value(corecontext.EnvironmentContextKey).(appmodule.Environment)
	if !ok {
		return nil, errors.New("environment not set")
	}

	headerInfo := environment.HeaderService.HeaderInfo(ctx)

	return &StatusResponse{
		Height:        uint64(headerInfo.Height),
		Timestamp:     &headerInfo.Time,
		AppHash:       headerInfo.AppHash,
		ValidatorHash: headerInfo.Hash,
	}, nil
}
