package errors

const (
	CodeTypeInternalErr        uint32 = 1
	CodeTypeEncodingErr        uint32 = 2
	CodeTypeUnauthorized       uint32 = 3
	CodeTypeUnknownRequest     uint32 = 4
	CodeTypeUnknownAddress     uint32 = 5
	CodeTypeBaseUnknownAddress uint32 = 5 // lol fuck it
	CodeTypeBadNonce           uint32 = 6

	CodeTypeBaseInvalidInput  uint32 = 20
	CodeTypeBaseInvalidOutput uint32 = 21
)
