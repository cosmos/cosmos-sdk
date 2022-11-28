package types

import errorsmod "cosmossdk.io/errors"

const StoreCodespace = "store"

var ErrInvalidProof = errorsmod.Register(StoreCodespace, 2, "invalid proof")

var (
	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = errorsmod.Register(StoreCodespace, 2, "tx parse error")

	// ErrInvalidSequence is used the sequence number (nonce) is incorrect
	// for the signature
	ErrInvalidSequence = errorsmod.Register(StoreCodespace, 3, "invalid sequence")

	// ErrUnauthorized is used whenever a request without sufficient
	// authorization is handled.
	ErrUnauthorized = errorsmod.Register(StoreCodespace, 4, "unauthorized")

	// ErrInsufficientFunds is used when the account cannot pay requested amount.
	ErrInsufficientFunds = errorsmod.Register(StoreCodespace, 5, "insufficient funds")

	// ErrUnknownRequest to doc
	ErrUnknownRequest = errorsmod.Register(StoreCodespace, 6, "unknown request")

	// ErrInvalidAddress to doc
	ErrInvalidAddress = errorsmod.Register(StoreCodespace, 7, "invalid address")

	// ErrInvalidPubKey to doc
	ErrInvalidPubKey = errorsmod.Register(StoreCodespace, 8, "invalid pubkey")

	// ErrUnknownAddress to doc
	ErrUnknownAddress = errorsmod.Register(StoreCodespace, 9, "unknown address")

	// ErrInvalidCoins to doc
	ErrInvalidCoins = errorsmod.Register(StoreCodespace, 10, "invalid coins")

	// ErrOutOfGas to doc
	ErrOutOfGas = errorsmod.Register(StoreCodespace, 11, "out of gas")

	// ErrMemoTooLarge to doc
	ErrMemoTooLarge = errorsmod.Register(StoreCodespace, 12, "memo too large")

	// ErrInsufficientFee to doc
	ErrInsufficientFee = errorsmod.Register(StoreCodespace, 13, "insufficient fee")

	// ErrTooManySignatures to doc
	ErrTooManySignatures = errorsmod.Register(StoreCodespace, 14, "maximum number of signatures exceeded")

	// ErrNoSignatures to doc
	ErrNoSignatures = errorsmod.Register(StoreCodespace, 15, "no signatures supplied")

	// ErrJSONMarshal defines an ABCI typed JSON marshalling error
	ErrJSONMarshal = errorsmod.Register(StoreCodespace, 16, "failed to marshal JSON bytes")

	// ErrJSONUnmarshal defines an ABCI typed JSON unmarshalling error
	ErrJSONUnmarshal = errorsmod.Register(StoreCodespace, 17, "failed to unmarshal JSON bytes")

	// ErrInvalidRequest defines an ABCI typed error where the request contains
	// invalid data.
	ErrInvalidRequest = errorsmod.Register(StoreCodespace, 18, "invalid request")

	// ErrTxInMempoolCache defines an ABCI typed error where a tx already exists
	// in the mempool.
	ErrTxInMempoolCache = errorsmod.Register(StoreCodespace, 19, "tx already in mempool")

	// ErrMempoolIsFull defines an ABCI typed error where the mempool is full.
	ErrMempoolIsFull = errorsmod.Register(StoreCodespace, 20, "mempool is full")

	// ErrTxTooLarge defines an ABCI typed error where tx is too large.
	ErrTxTooLarge = errorsmod.Register(StoreCodespace, 21, "tx too large")

	// ErrKeyNotFound defines an error when the key doesn't exist
	ErrKeyNotFound = errorsmod.Register(StoreCodespace, 22, "key not found")

	// ErrWrongPassword defines an error when the key password is invalid.
	ErrWrongPassword = errorsmod.Register(StoreCodespace, 23, "invalid account password")

	// ErrorInvalidSigner defines an error when the tx intended signer does not match the given signer.
	ErrorInvalidSigner = errorsmod.Register(StoreCodespace, 24, "tx intended signer does not match the given signer")

	// ErrorInvalidGasAdjustment defines an error for an invalid gas adjustment
	ErrorInvalidGasAdjustment = errorsmod.Register(StoreCodespace, 25, "invalid gas adjustment")

	// ErrInvalidHeight defines an error for an invalid height
	ErrInvalidHeight = errorsmod.Register(StoreCodespace, 26, "invalid height")

	// ErrInvalidVersion defines a general error for an invalid version
	ErrInvalidVersion = errorsmod.Register(StoreCodespace, 27, "invalid version")

	// ErrInvalidChainID defines an error when the chain-id is invalid.
	ErrInvalidChainID = errorsmod.Register(StoreCodespace, 28, "invalid chain-id")

	// ErrInvalidType defines an error an invalid type.
	ErrInvalidType = errorsmod.Register(StoreCodespace, 29, "invalid type")

	// ErrTxTimeoutHeight defines an error for when a tx is rejected out due to an
	// explicitly set timeout height.
	ErrTxTimeoutHeight = errorsmod.Register(StoreCodespace, 30, "tx timeout height")

	// ErrUnknownExtensionOptions defines an error for unknown extension options.
	ErrUnknownExtensionOptions = errorsmod.Register(StoreCodespace, 31, "unknown extension options")

	// ErrWrongSequence defines an error where the account sequence defined in
	// the signer info doesn't match the account's actual sequence number.
	ErrWrongSequence = errorsmod.Register(StoreCodespace, 32, "incorrect account sequence")

	// ErrPackAny defines an error when packing a protobuf message to Any fails.
	ErrPackAny = errorsmod.Register(StoreCodespace, 33, "failed packing protobuf message to Any")

	// ErrUnpackAny defines an error when unpacking a protobuf message from Any fails.
	ErrUnpackAny = errorsmod.Register(StoreCodespace, 34, "failed unpacking protobuf message from Any")

	// ErrLogic defines an internal logic error, e.g. an invariant or assertion
	// that is violated. It is a programmer error, not a user-facing error.
	ErrLogic = errorsmod.Register(StoreCodespace, 35, "internal logic error")

	// ErrConflict defines a conflict error, e.g. when two goroutines try to access
	// the same resource and one of them fails.
	ErrConflict = errorsmod.Register(StoreCodespace, 36, "conflict")

	// ErrNotSupported is returned when we call a branch of a code which is currently not
	// supported.
	ErrNotSupported = errorsmod.Register(StoreCodespace, 37, "feature not supported")

	// ErrNotFound defines an error when requested entity doesn't exist in the state.
	ErrNotFound = errorsmod.Register(StoreCodespace, 38, "not found")

	// ErrIO should be used to wrap internal errors caused by external operation.
	// Examples: not DB domain error, file writing etc...
	ErrIO = errorsmod.Register(StoreCodespace, 39, "Internal IO error")

	// ErrAppConfig defines an error occurred if min-gas-prices field in BaseConfig is empty.
	ErrAppConfig = errorsmod.Register(StoreCodespace, 40, "error in app.toml")

	// ErrInvalidGasLimit defines an error when an invalid GasWanted value is
	// supplied.
	ErrInvalidGasLimit = errorsmod.Register(StoreCodespace, 41, "invalid gas limit")

	// ErrPanic should only be set when we recovering from a panic
	ErrPanic = errorsmod.ErrPanic
)
