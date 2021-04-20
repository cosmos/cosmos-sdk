<!--
order: 1
-->

# State

## ConstantFee

Due to the anticipated large gas cost requirement to verify an invariant (and
potential to exceed the maximum allowable block gas limit) a constant fee is
used instead of the standard gas consumption method. The constant fee is
intended to be larger than the anticipated gas cost of running the invariant
with the standard gas consumption method. 

The ConstantFee param is held in the global params store. 

 - Params: `mint/params -> legacy_amino(sdk.Coin)`

