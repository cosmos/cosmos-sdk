package watchers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"cosmossdk.io/log"
)

// blockEndpoints lists the RPC endpoints to try for getting the latest block height
var blockEndpoints = []string{"/v1/block", "/block"}

func NewHTTPRPCBLockChecker(baseUrl string, logger log.Logger) HeightChecker {
	return &httpRPCBlockChecker{
		baseUrl: baseUrl,
		logger:  logger,
	}
}

type httpRPCBlockChecker struct {
	baseUrl string
	subUrl  string
	logger  log.Logger
}

func (j *httpRPCBlockChecker) GetLatestBlockHeight() (uint64, error) {
	if j.subUrl != "" {
		return j.getLatestBlockHeight(j.subUrl)
	}

	var errs []error
	for _, endpoint := range blockEndpoints {
		height, err := j.getLatestBlockHeight(endpoint)
		if err == nil {
			j.logger.Info("Successfully resolved latest block height", "url", j.baseUrl+endpoint)
			j.subUrl = endpoint
			return height, nil
		}
		errs = append(errs, err)
	}

	return 0, fmt.Errorf("failed to get latest block height from RPC endpoints %v: %w", blockEndpoints, errors.Join(errs...))
}

func (j *httpRPCBlockChecker) getLatestBlockHeight(subUrl string) (uint64, error) {
	url := j.baseUrl + subUrl
	res, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block height: %w", err)
	}
	defer res.Body.Close()

	bz, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read latest block height: %w", err)
	}

	return getHeightFromRPCBlockResponse(bz)
}

var _ HeightChecker = &httpRPCBlockChecker{}

type Header struct {
	Height string `json:"height"`
}
type Block struct {
	Header Header `json:"header"`
}
type Result struct {
	Block Block `json:"block"`
}
type Response struct {
	Result Result `json:"result"`
}

func getHeightFromRPCBlockResponse(bz []byte) (uint64, error) {
	var response Response
	err := json.Unmarshal(bz, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal block response: %w", err)
	}

	height := response.Result.Block.Header.Height
	return strconv.ParseUint(height, 10, 64)
}
