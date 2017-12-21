package coin

type CoinHolder interface {
	GetCoins() Coins
	SetCoins(Coins)
}
