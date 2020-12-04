package tendermint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TxResponse struct {
	Hash string `json:"hash"`
}

func (c Client) Tx(hash string) (TxResponse, error) {
	resp, err := http.Get(c.getEndpoint(fmt.Sprintf("tx?hash=%s", hash)))
	if err != nil {
		return TxResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TxResponse{}, err
	}

	var jsonResp map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return TxResponse{}, err
	}

	var blockResponse TxResponse
	err = json.Unmarshal(jsonResp["result"], &blockResponse)
	if err != nil {
		return TxResponse{}, err
	}

	return blockResponse, nil
}
