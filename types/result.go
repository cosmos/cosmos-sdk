package types

import (
	"fmt"
	"strings"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Result is the union of ResponseFormat and ResponseCheckTx.
type Result struct {

	// Code is the response code, is stored back on the chain.
	Code CodeType

	// Codespace is the string referring to the domain of an error
	Codespace CodespaceType

	// Data is any data returned from the app.
	// Data has to be length prefixed in order to separate
	// results from multiple msgs executions
	Data []byte

	// Log is just debug information. NOTE: nondeterministic.
	Log string

	// GasWanted is the maximum units of work we allow this tx to perform.
	GasWanted uint64

	// GasUsed is the amount of gas actually consumed. NOTE: unimplemented
	GasUsed uint64

	// Tags are used for transaction indexing and pubsub.
	Tags Tags
}

// TODO: In the future, more codes may be OK.
func (res Result) IsOK() bool {
	return res.Code.IsOK()
}

// Is a version of TxResponse where the tags are StringTags rather than []byte tags
type TxResponse struct {
	Height    int64      `json:"height"`
	TxHash    string     `json:"txhash"`
	Code      uint32     `json:"code,omitempty"`
	Data      []byte     `json:"data,omitempty"`
	Log       string     `json:"log,omitempty"`
	Info      string     `json:"info,omitempty"`
	GasWanted int64      `json:"gas_wanted,omitempty"`
	GasUsed   int64      `json:"gas_used,omitempty"`
	Tags      StringTags `json:"tags,omitempty"`
	Codespace string     `json:"codespace,omitempty"`
	Tx        Tx         `json:"tx,omitempty"`
}

// NewResponseResultTx returns a TxResponse given a ResultTx from tendermint
func NewResponseResultTx(res *ctypes.ResultTx, tx Tx) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	return TxResponse{
		TxHash:    res.Hash.String(),
		Height:    res.Height,
		Code:      res.TxResult.Code,
		Data:      res.TxResult.Data,
		Log:       res.TxResult.Log,
		Info:      res.TxResult.Info,
		GasWanted: res.TxResult.GasWanted,
		GasUsed:   res.TxResult.GasUsed,
		Tags:      TagsToStringTags(res.TxResult.Tags),
		Tx:        tx,
	}
}

// NewResponseFormatBroadcastTxCommit returns a TxResponse given a ResultBroadcastTxCommit from tendermint
func NewResponseFormatBroadcastTxCommit(res *ctypes.ResultBroadcastTxCommit) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	return TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Code:      res.DeliverTx.Code,
		Data:      res.DeliverTx.Data,
		Log:       res.DeliverTx.Log,
		Info:      res.DeliverTx.Info,
		GasWanted: res.DeliverTx.GasWanted,
		GasUsed:   res.DeliverTx.GasUsed,
		Tags:      TagsToStringTags(res.DeliverTx.Tags),
		Codespace: res.DeliverTx.Codespace,
	}

}

// NewResponseFormatBroadcastTx returns a TxResponse given a ResultBroadcastTx from tendermint
func NewResponseFormatBroadcastTx(res *ctypes.ResultBroadcastTx) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	return TxResponse{
		Code:   res.Code,
		Data:   res.Data.Bytes(),
		Log:    res.Log,
		TxHash: res.Hash.String(),
	}
}

func (r TxResponse) String() string {
	var sb strings.Builder
	sb.WriteString("Response:\n")

	if r.Height > 0 {
		sb.WriteString(fmt.Sprintf("  Height: %d\n", r.Height))
	}

	if r.TxHash != "" {
		sb.WriteString(fmt.Sprintf("  TxHash: %s\n", r.TxHash))
	}

	if r.Code > 0 {
		sb.WriteString(fmt.Sprintf("  Code: %d\n", r.Code))
	}

	if r.Data != nil {
		sb.WriteString(fmt.Sprintf("  Data: %s\n", string(r.Data)))
	}

	if r.Log != "" {
		sb.WriteString(fmt.Sprintf("  Log: %s\n", r.Log))
	}

	if r.Info != "" {
		sb.WriteString(fmt.Sprintf("  Info: %s\n", r.Info))
	}

	if r.GasWanted != 0 {
		sb.WriteString(fmt.Sprintf("  GasWanted: %d\n", r.GasWanted))
	}

	if r.GasUsed != 0 {
		sb.WriteString(fmt.Sprintf("  GasUsed: %d\n", r.GasUsed))
	}

	if len(r.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("  Tags: \n%s\n", r.Tags.String()))
	}

	if r.Codespace != "" {
		sb.WriteString(fmt.Sprintf("  Codespace: %s\n", r.Codespace))
	}

	return strings.TrimSpace(sb.String())
}

// Empty returns true if the response is empty
func (r TxResponse) Empty() bool {
	return r.TxHash == "" && r.Log == ""
}
