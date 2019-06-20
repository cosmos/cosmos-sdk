# Future improvements

The current supply module only keeps track of the `Total` supply of coins held in
the network. This module might be updated in the future to a track all the
[money supply](https://en.wikipedia.org/wiki/Money_supply) hold within a chain.
That design would have to implement the different types of "**M**"s but applied
to Cosmos' BPoS.

Future improvements may also include other types of supply such as:

* **Liquid Supply:** Supply of the network that is not locked. Coins locked can
  be delegated, undelegating or vesting (as part of a `VestingAccount`).
* **Vesting Supply:** Supply of the network that is part of a continuous or a
  delayed vesting account.
