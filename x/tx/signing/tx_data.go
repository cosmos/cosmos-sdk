package signing

import "google.golang.org/protobuf/types/known/timestamppb"

// RawMsg is a raw protobuf message represented as TypeUrl + marshaled Value bytes.
// It has the same structure as google.protobuf.Any but avoids import cycles by
// not depending on any SDK module that transitively imports x/tx/signing.
type RawMsg struct {
	TypeUrl string
	Value   []byte
}

// TxBodyData holds the body fields needed by sign mode handlers.
type TxBodyData struct {
	Messages                    []RawMsg
	Memo                        string
	TimeoutHeight               uint64
	Unordered                   bool
	TimeoutTimestamp            *timestamppb.Timestamp
	ExtensionOptions            []RawMsg
	NonCriticalExtensionOptions []RawMsg
}

// TxAuthInfoData holds the auth info fields needed by sign mode handlers.
// SignerInfos is omitted — no handler reads it; the signer addresses are
// pre-computed in TxData.Signers.
type TxAuthInfoData struct {
	Fee TxFeeData
}

// TxFeeData holds fee data used in amino signing.
type TxFeeData struct {
	Amount   []TxCoinData
	GasLimit uint64
	Payer    string
	Granter  string
}

// TxCoinData is a denomination/amount pair for fee display in amino signing.
type TxCoinData struct {
	Denom  string
	Amount string
}

// TxData is the data about a transaction necessary to generate sign bytes.
type TxData struct {
	Body     *TxBodyData
	AuthInfo *TxAuthInfoData

	// BodyBytes is the marshaled body bytes that will be part of TxRaw.
	BodyBytes []byte

	// AuthInfoBytes is the marshaled AuthInfo bytes that will be part of TxRaw.
	AuthInfoBytes []byte

	// BodyHasUnknownNonCriticals should be set to true if the transaction has been
	// decoded and found to have unknown non-critical fields. This is only needed
	// for amino JSON signing.
	BodyHasUnknownNonCriticals bool
}
