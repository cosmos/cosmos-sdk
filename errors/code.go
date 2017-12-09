package errors

const (
	defaultErrCode uint32 = 0x1

	CodeTypeInternalErr        uint32 = 0
	CodeTypeEncodingErr        uint32 = 1
	CodeTypeUnauthorized       uint32 = 2
	CodeTypeUnknownRequest     uint32 = 3
	CodeTypeUnknownAddress     uint32 = 4
	CodeTypeBaseUnknownAddress uint32 = 4 // lol fuck it
	CodeTypeBadNonce           uint32 = 5
	CodeTypeBaseInvalidInput   uint32 = 20
	CodeTypeBaseInvalidOutput  uint32 = 21
)
