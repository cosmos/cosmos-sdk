```golang
type SubKeyAccount struct {
  Address       AccAddress
  Coins         Coins
  PubKey        PubKey
  AccountNumber uint64
  Sequence      uint64
  SubKeys       map[key]
}
```

```golang
type SubKeyMetadata struct {
  PermissionedRoutes   []string
  DailyFeeAllowance    sdk.Coins
  DailyFeeUsed         sdk.Coins
}
```

SubKey DailyFee Window Queue
Similar to Queues used in staking and gov.

```golang
type DailyFeeSpend struct {
  Address              sdk.AccAddress
  SubKey               sdk.AccPubKey
  FeeSpent             sdk.Coins
}
```

