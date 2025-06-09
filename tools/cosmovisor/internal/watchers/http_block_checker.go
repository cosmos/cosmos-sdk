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

func NewHTTPRPCBLockChecker(baseUrl string, logger log.Logger) HeightChecker {
	return httpRPCBlockChecker{
		baseUrl: baseUrl,
		logger:  logger,
	}
}

type httpRPCBlockChecker struct {
	baseUrl string
	subUrl  string
	logger  log.Logger
}

func (j httpRPCBlockChecker) GetLatestBlockHeight() (uint64, error) {
	if j.subUrl != "" {
		return j.getLatestBlockHeight(j.subUrl)
	}

	height, err1 := j.getLatestBlockHeight("/v1/block")
	if err1 == nil {
		j.logger.Info("Successfully resolved latest block height from /v1/block", "url", j.baseUrl+"/v1/block")
		// If we successfully got the height from /v1/block, we can cache the subUrl
		j.subUrl = "/v1/block"
		return height, nil
	}

	height, err2 := j.getLatestBlockHeight("/block")
	if err2 == nil {
		j.logger.Info("Successfully resolved latest block height from /block", "url", j.baseUrl+"/block")
		// If we successfully got the height from /block, we can cache the subUrl
		j.subUrl = "/block"
		return height, nil
	}

	return 0, fmt.Errorf("failed to get latest block height from both /block and /v1/block RPC endpoints: %w", errors.Join(err1, err2))
}

func (j httpRPCBlockChecker) getLatestBlockHeight(subUrl string) (uint64, error) {
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

var _ HeightChecker = httpRPCBlockChecker{}

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
