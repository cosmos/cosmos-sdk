package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (c Client) GetTx(ctx context.Context, hash string) (sdk.TxResponse, error) {
	path := fmt.Sprintf("/txs/%s", hash)

	r, err := http.Get(c.getEndpoint(path))
	if err != nil {
		return sdk.TxResponse{}, err
	}
	if r == nil {
		return sdk.TxResponse{}, fmt.Errorf("unable to fetch data from endpoint %s", c.getEndpoint(path))
	}
	defer r.Body.Close()
	btes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	var txRes sdk.TxResponse
	if err = c.cdc.UnmarshalJSON(btes, &txRes); err != nil {
		return sdk.TxResponse{}, err
	}

	return txRes, nil
}

func (c Client) PostTx(ctx context.Context, bz []byte) (sdk.TxResponse, error) {
	r, err := http.Post(c.getEndpoint("/txs"), "application/json", bytes.NewBuffer(bz))
	if err != nil {
		return sdk.TxResponse{}, err
	}
	if r == nil {
		return sdk.TxResponse{}, fmt.Errorf("unable to get response from endpoint %s", c.getEndpoint("/txs"))
	}
	defer r.Body.Close()
	btes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	var txRes sdk.TxResponse
	if err = c.cdc.UnmarshalJSON(btes, &txRes); err != nil {
		return sdk.TxResponse{}, err
	}

	return txRes, nil
}
