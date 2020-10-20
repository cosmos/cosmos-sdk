package tendermint

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type StatusResponse struct {
	NodeInfo StatusNodeInfo `json:"node_info"`
}

type StatusNodeInfo struct {
	Network string `json:"network"`
}

func (c Client) Status() (StatusResponse, error) {
	resp, err := http.Get(c.getEndpoint("status"))
	if err != nil {
		return StatusResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StatusResponse{}, err
	}
	var jsonResp map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return StatusResponse{}, err
	}

	var statusResp StatusResponse
	err = json.Unmarshal(jsonResp["result"], &statusResp)
	if err != nil {
		return StatusResponse{}, err
	}

	return statusResp, err
}
