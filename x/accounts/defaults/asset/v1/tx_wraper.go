package v1

// MsgInitAssetAccountWrapper wrap MsgInitAssetAccount with
// send, mint, burn func logic
type MsgInitAssetAccountWrapper struct {
	MsgInitAssetAccount
	TransferFunc func(aa AssetAccountI) SendFunc
	MintFunc     func(aa AssetAccountI) MintFunc
	BurnFunc     func(aa AssetAccountI) BurnFunc
}
