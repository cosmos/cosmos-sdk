package sdk

import (
	"context"

	"github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/types"
)

func (c Client) GetAuthAccount(ctx context.Context, address string, height int64) (types.AccountResponse, error) {
	//var accRes rest.ResponseWithHeight
	//path := fmt.Sprintf("/auth/accounts/%s?height=%d", address, height)
	//r, err := http.Get(c.getEndpoint(path))
	//if err != nil {
	//	return types.AccountResponse{}, err
	//}
	//if r == nil {
	//	return types.AccountResponse{}, fmt.Errorf("unable to fetch data from endpoint %s", c.getEndpoint(path))
	//}
	//btes, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	return types.AccountResponse{}, err
	//}
	//
	//if err = c.cdc.UnmarshalBinaryBare(btes, &accRes); err != nil {
	//	return types.AccountResponse{}, err
	//}
	//
	//var res types.Response
	//err = json.Unmarshal(accRes.Result, &res)
	//if err != nil {
	//	return types.AccountResponse{}, err
	//}
	//defer r.Body.Close()
	//
	//return types.AccountResponse{
	//	Height: accRes.Height,
	//	Result: res,
	//}, nil
	return types.AccountResponse{}, nil
}
