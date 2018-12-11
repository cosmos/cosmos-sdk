package types

type AccountStoreCache interface {
	GetAccount(addr AccAddress) Account
	SetAccount(addr AccAddress, acc Account)
	Delete(addr AccAddress)
}

type AccountCache interface {
	AccountStoreCache

	Cache() AccountCache
	Write()
}
