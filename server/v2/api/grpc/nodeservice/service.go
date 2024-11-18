package nodeservice

import (
	context "context"
	fmt "fmt"

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

func (s queryServer) Status(ctx context.Context, _ *StatusRequest) (*StatusResponse, error) {
	// note, environment nor execution context isn't available in the context

	// return &StatusResponse{
	// 	Height:        uint64(headerInfo.Height),
	// 	Timestamp:     &headerInfo.Time,
	// 	AppHash:       headerInfo.AppHash,
	// 	ValidatorHash: headerInfo.Hash,
	// }, nil

	return &StatusResponse{}, fmt.Errorf("not implemented")
}
