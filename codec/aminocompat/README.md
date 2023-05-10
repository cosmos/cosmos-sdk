## aminocompat

The purpose of this code is to help with checking the logic of compatibility with go-amino before
apply any optimizations.

### Uses
For example as per [issue #2350](https://github.com/cosmos/cosmos-sdk/issues/2350) in which it was deemed
that the cosmos-sdk invoked amino.MarshalJSON then encoding/json.Unmarshal then encoding/json.Marshal so as
to just get sorted JSON. An idea was floated to perhaps head directly to using encoding/json.Marshal

The tricky thing is that skipping amino checks would mean that previously unsupported types like floating points,
complex numbers, enums, maps would now be blindly supported, yet the module requires an explicit `.Unsafe=true` to
be set.

To solve that problem, we've implemented a pass `AllClear` which checks that a value's contents are amino-compatible
and if all clear, permits us to use encoding/json.Marshal directly with stark improvements when used with MsgDeposit:

```shell
$ benchstat before.txt after.txt 
name          old time/op    new time/op    delta
MsgDeposit-8    10.8µs ± 1%     1.3µs ± 1%  -87.55%  (p=0.000 n=10+8)

name          old alloc/op   new alloc/op   delta
MsgDeposit-8    3.75kB ± 0%    0.26kB ± 0%  -92.96%  (p=0.000 n=10+10)

name          old allocs/op  new allocs/op  delta
MsgDeposit-8      99.0 ± 0%      13.0 ± 0%  -86.87%  (p=0.000 n=10+10)
```
