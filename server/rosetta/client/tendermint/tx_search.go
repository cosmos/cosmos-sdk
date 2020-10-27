package tendermint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TxSearchResponse struct {
	Txs        []TxSearchResponseResultTxs `json:"txs"`
	TotalCount string                      `json:"total_count"`
}

type TxSearchResponseResultTxs struct {
	Hash string `json:"hash,omitempty"`
}

func (c Client) TxSearch(query string) (TxSearchResponse, error) {
	resp, err := http.Get(c.getEndpoint(fmt.Sprintf("tx_search?query=\"%s\"", query)))
	if err != nil {
		return TxSearchResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TxSearchResponse{}, err
	}

	var jsonResp map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return TxSearchResponse{}, err
	}

	var txSearchResponse TxSearchResponse
	err = json.Unmarshal(jsonResp["result"], &txSearchResponse)
	if err != nil {
		return TxSearchResponse{}, err
	}

	return txSearchResponse, nil
}
