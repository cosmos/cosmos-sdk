## State

### Inflation
 - key: `0x00`
 - value: `amino(Inflation)`

The current annual inflation rate.

```golang
type Inflation sdk.Rat 
```

### InflationLastTime
 - key: `0x01`
 - value: `amino(InflationLastTime)`

The last unix time which the inflation was processed for. 

```golang
type InflationLastTime int64
```
