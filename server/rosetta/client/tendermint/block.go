package tendermint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type BlockResponse struct {
	BlockID BlockID `json:"block_id,omitempty"`
	Block   Block   `json:"block,omitempty"`
}

type BlockID struct {
	Hash string `json:"hash"`
}

type Block struct {
	Header BlockHeader `json:"header,omitempty"`
}

type BlockHeader struct {
	LastBlockID BlockID `json:"last_block_id"`
	Height      string  `json:"height"`
	Time        string  `json:"time"`
}

func (c Client) Block(height uint64) (BlockResponse, error) {
	var endpoint string
	if height == 0 {
		endpoint = c.getEndpoint("block")
	} else {
		endpoint = c.getEndpoint(fmt.Sprintf("block?height=%d", height))
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		return BlockResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BlockResponse{}, err
	}

	var jsonResp map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return BlockResponse{}, err
	}

	var blockResponse BlockResponse
	err = json.Unmarshal(jsonResp["result"], &blockResponse)
	if err != nil {
		return BlockResponse{}, err
	}

	return blockResponse, nil
}

func (c Client) BlockByHash(hash string) (BlockResponse, error) {
	resp, err := http.Get(c.getEndpoint(fmt.Sprintf("block_by_hash?hash=%s", hash)))
	if err != nil {
		return BlockResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BlockResponse{}, err
	}

	var jsonResp map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return BlockResponse{}, err
	}

	var blockResponse BlockResponse
	err = json.Unmarshal(jsonResp["result"], &blockResponse)
	if err != nil {
		return BlockResponse{}, err
	}

	return blockResponse, nil
}
