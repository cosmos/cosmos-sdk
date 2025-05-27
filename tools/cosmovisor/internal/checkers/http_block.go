package checkers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func NewHTTPRPCBLockChecker(url string) LatestBlockChecker {
	panic("implement me")
}

type httpRPCBlockChecker struct {
	url string
}

func (j httpRPCBlockChecker) GetLatestBlockHeight() (uint64, error) {
	res, err := http.Get(j.url)
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

var _ LatestBlockChecker = httpRPCBlockChecker{}

func getHeightFromRPCBlockResponse(bz []byte) (uint64, error) {
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

	var response Response
	err := json.Unmarshal(bz, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal block response: %w", err)
	}

	height := response.Result.Block.Header.Height
	return strconv.ParseUint(height, 10, 64)
}
