package types

import (
	"encoding/hex"
	"encoding/json"
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

	// Log contains the txs log information. NOTE: nondeterministic.
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

// ABCIMessageLogs represents a slice of ABCIMessageLog.
type ABCIMessageLogs []ABCIMessageLog

// ABCIMessageLog defines a structure containing an indexed tx ABCI message log.
type ABCIMessageLog struct {
	MsgIndex int    `json:"msg_index"`
	Success  bool   `json:"success"`
	Log      string `json:"log"`
}

// String implements the fmt.Stringer interface for the ABCIMessageLogs type.
func (logs ABCIMessageLogs) String() (str string) {
	if logs != nil {
		raw, err := json.Marshal(logs)
		if err == nil {
			str = string(raw)
		}
	}

	return str
}

// TxResponse defines a structure containing relevant tx data and metadata. The
// tags are stringified and the log is JSON decoded.
type TxResponse struct {
	Height    int64           `json:"height"`
	TxHash    string          `json:"txhash"`
	Code      uint32          `json:"code,omitempty"`
	Data      string          `json:"data,omitempty"`
	RawLog    string          `json:"raw_log,omitempty"`
	Logs      ABCIMessageLogs `json:"logs,omitempty"`
	Info      string          `json:"info,omitempty"`
	GasWanted int64           `json:"gas_wanted,omitempty"`
	GasUsed   int64           `json:"gas_used,omitempty"`
	Tags      StringTags      `json:"tags,omitempty"`
	Codespace string          `json:"codespace,omitempty"`
	Tx        Tx              `json:"tx,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

// NewResponseResultTx returns a TxResponse given a ResultTx from tendermint
func NewResponseResultTx(res *ctypes.ResultTx, tx Tx, timestamp string) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	parsedLogs, _ := ParseABCILogs(res.TxResult.Log)

	return TxResponse{
		TxHash:    res.Hash.String(),
		Height:    res.Height,
		Code:      res.TxResult.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.TxResult.Data)),
		RawLog:    res.TxResult.Log,
		Logs:      parsedLogs,
		Info:      res.TxResult.Info,
		GasWanted: res.TxResult.GasWanted,
		GasUsed:   res.TxResult.GasUsed,
		Tags:      TagsToStringTags(res.TxResult.Tags),
		Tx:        tx,
		Timestamp: timestamp,
	}
}

// NewResponseFormatBroadcastTxCommit returns a TxResponse given a
// ResultBroadcastTxCommit from tendermint.
func NewResponseFormatBroadcastTxCommit(res *ctypes.ResultBroadcastTxCommit) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	if !res.CheckTx.IsOK() {
		return newTxResponseCheckTx(res)
	}

	return newTxResponseDeliverTx(res)
}

func newTxResponseCheckTx(res *ctypes.ResultBroadcastTxCommit) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := ParseABCILogs(res.CheckTx.Log)

	return TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Code:      res.CheckTx.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.CheckTx.Data)),
		RawLog:    res.CheckTx.Log,
		Logs:      parsedLogs,
		Info:      res.CheckTx.Info,
		GasWanted: res.CheckTx.GasWanted,
		GasUsed:   res.CheckTx.GasUsed,
		Tags:      TagsToStringTags(res.CheckTx.Tags),
		Codespace: res.CheckTx.Codespace,
	}
}

func newTxResponseDeliverTx(res *ctypes.ResultBroadcastTxCommit) TxResponse {
	if res == nil {
		return TxResponse{}
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := ParseABCILogs(res.DeliverTx.Log)

	return TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Code:      res.DeliverTx.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.DeliverTx.Data)),
		RawLog:    res.DeliverTx.Log,
		Logs:      parsedLogs,
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

	parsedLogs, _ := ParseABCILogs(res.Log)

	return TxResponse{
		Code:   res.Code,
		Data:   res.Data.String(),
		RawLog: res.Log,
		Logs:   parsedLogs,
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

	if r.Data != "" {
		sb.WriteString(fmt.Sprintf("  Data: %s\n", r.Data))
	}

	if r.RawLog != "" {
		sb.WriteString(fmt.Sprintf("  Raw Log: %s\n", r.RawLog))
	}

	if r.Logs != nil {
		sb.WriteString(fmt.Sprintf("  Logs: %s\n", r.Logs))
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

	if r.Timestamp != "" {
		sb.WriteString(fmt.Sprintf("  Timestamp: %s\n", r.Timestamp))
	}

	return strings.TrimSpace(sb.String())
}

// Empty returns true if the response is empty
func (r TxResponse) Empty() bool {
	return r.TxHash == "" && r.Logs == nil
}

// ParseABCILogs attempts to parse a stringified ABCI tx log into a slice of
// ABCIMessageLog types. It returns an error upon JSON decoding failure.
func ParseABCILogs(logs string) (res ABCIMessageLogs, err error) {
	err = json.Unmarshal([]byte(logs), &res)
	return res, err
}
