package rest

import "github.com/cosmos/cosmos-sdk/x/auth"

// BroadcastReq requests broadcasting a transaction
type BroadcastReq struct {
	Tx     auth.StdTx `json:"tx"`
	Return string     `json:"return"`
}
