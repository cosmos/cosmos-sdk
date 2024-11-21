package nodeservice

import (
	context "context"

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
	minGasPrices, ok := s.cfg["server"].(map[string]interface{})["minimum-gas-prices"]
	if ok {
		minGasPricesStr = minGasPrices.(string)
	}

	return &ConfigResponse{
		MinimumGasPrice: minGasPricesStr,
	}, nil
}
