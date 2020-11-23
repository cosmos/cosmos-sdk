package types

import "strconv"

const (
	// SubModuleName defines the IBC client name
	SubModuleName string = "client"

	// RouterKey is the message route for IBC client
	RouterKey string = SubModuleName

	// QuerierRoute is the querier route for IBC client
	QuerierRoute string = SubModuleName
)

var (
	// HistoricalInfoKeyPrefix is the store key prefix for the historical info
	HistoricalInfoKeyPrefix = []byte("historicalInfo")
)

// GetHistoricalInfoKey returns the key for indexing HistoricalInfo objects.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKeyPrefix, []byte(strconv.FormatInt(height, 10))...)
}
