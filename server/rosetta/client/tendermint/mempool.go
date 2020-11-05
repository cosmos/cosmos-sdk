package tendermint

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type UnconfirmedTxsResponse struct {
	Txs []string `json:"txs"`
}

func (c Client) UnconfirmedTxs() (UnconfirmedTxsResponse, error) {
	resp, err := http.Get(c.getEndpoint("unconfirmed_txs"))
	if err != nil {
		return UnconfirmedTxsResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return UnconfirmedTxsResponse{}, err
	}

	var jsonResp map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return UnconfirmedTxsResponse{}, err
	}

	var unconfirmedTxsResp UnconfirmedTxsResponse
	err = json.Unmarshal(jsonResp["result"], &unconfirmedTxsResp)
	if err != nil {
		return UnconfirmedTxsResponse{}, err
	}

	return unconfirmedTxsResp, nil
}
