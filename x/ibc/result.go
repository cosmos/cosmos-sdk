package ibc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC handler returns OK if the inner transaction executed
// no matter it succeed or not
// so store the code in the data field

type ResultData struct {
	Code sdk.ABCICodeType
	Data []byte
}

func WrapResult(res1 sdk.Result) (res2 sdk.Result) {
	res2 = res1
	res2.Code = sdk.ABCICodeOK
	res2.Data = ResultData{res1.Code, res1.Data}.Marshal()
	return
}

func UnwrapResult(res1 sdk.Result) (res2 sdk.Result) {
	res2 = res1
	var data ResultData
	if (&data).Unmarshal(res1.Data) != nil {
		return res1
	}
	res2.Code = data.Code
	res2.Data = data.Data
	return
}

func (data ResultData) Marshal() []byte {
	bz, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return bz
}

func (data *ResultData) Unmarshal(bz []byte) error {
	return json.Unmarshal(bz, data)
}
