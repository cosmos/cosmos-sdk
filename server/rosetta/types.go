package rosetta

import (
	"crypto/sha256"
)

// statuses
const (
	StatusTxSuccess   = "Success"
	StatusTxReverted  = "Reverted"
	StatusPeerSynced  = "synced"
	StatusPeerSyncing = "syncing"
)
const (
	DeliverTxSize           = sha256.Size
	BeginEndBlockTxSize     = DeliverTxSize + 1
	EndBlockHashStart       = 0x0
	BeginBlockHashStart     = 0x1
	BurnerAddressIdentifier = "burner"
)

// TransactionType is used to distinguish if a rosetta provided hash
// represents endblock, beginblock or deliver tx
type TransactionType int

const (
	BeginBlockTx TransactionType = iota
	EndBlockTx
	DeliverTxTx
	UnrecognizedTx
)

// metadata options

// misc
const (
	Log = "log"
)

// ConstructionPreprocessMetadata is used to represent
// the metadata rosetta can provide during preprocess options
type ConstructionPreprocessMetadata struct {
	Memo     string `json:"memo"`
	GasLimit uint64 `json:"gas_limit"`
	GasPrice string `json:"gas_price"`
}

func (c *ConstructionPreprocessMetadata) FromMetadata(meta map[string]interface{}) error {
	return unmarshalMetadata(meta, c)
}

// PreprocessOperationsOptionsResponse is the structured metadata options returned by the preprocess operations endpoint
type PreprocessOperationsOptionsResponse struct {
	ExpectedSigners []string `json:"expected_signers"`
	Memo            string   `json:"memo"`
	GasLimit        uint64   `json:"gas_limit"`
	GasPrice        string   `json:"gas_price"`
}

func (c PreprocessOperationsOptionsResponse) ToMetadata() (map[string]interface{}, error) {
	return marshalMetadata(c)
}

func (c *PreprocessOperationsOptionsResponse) FromMetadata(meta map[string]interface{}) error {
	return unmarshalMetadata(meta, c)
}

// SignerData contains information on the signers when the request
// is being created, used to populate the account information
type SignerData struct {
	AccountNumber uint64 `json:"account_number"`
	Sequence      uint64 `json:"sequence"`
}

// ConstructionMetadata are the metadata options used to
// construct a transaction. It is returned by ConstructionMetadataFromOptions
// and fed to ConstructionPayload to process the bytes to sign.
type ConstructionMetadata struct {
	ChainID     string        `json:"chain_id"`
	SignersData []*SignerData `json:"signer_data"`
	GasLimit    uint64        `json:"gas_limit"`
	GasPrice    string        `json:"gas_price"`
	Memo        string        `json:"memo"`
}

func (c ConstructionMetadata) ToMetadata() (map[string]interface{}, error) {
	return marshalMetadata(c)
}

func (c *ConstructionMetadata) FromMetadata(meta map[string]interface{}) error {
	return unmarshalMetadata(meta, c)
}
