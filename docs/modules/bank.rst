Bank
====

Defines how coins (i.e cryptocurrencies) are transferred.


::

    type MsgSend struct {
      Inputs  []Input  `json:"inputs"`
      Outputs []Output `json:"outputs"`
    }

The Input and Output structs are define an ``Address`` and a set of ``Coins``.

::

  type Input struct {
	   Address sdk.Address `json:"address"`
	   Coins   sdk.Coins   `json:"coins"`
  }

::

  type Output struct {
	   Address sdk.Address `json:"address"`
	   Coins   sdk.Coins   `json:"coins"`
  }

Keeper
------

A Bank ``Keeper`` is basically an ``AccountMapper`` that manages transfers between accounts.

There are 3 types of bank keepers:

- ``ViewKeeper``: only allows reading of balances
- ``SendKeeper``: only allows transfers between accounts, without the possibility of creating coins
- ``Keeper``: all the above, plus allowing the creation and deletion of coins to an ``Address``

::

    type Keeper struct {
      am sdk.AccountMapper
    }



Methods

::

    NewKeeper(am sdk.AccountMapper)

Returns a new ``Keeper``.

::

    keeper.GetCoins(ctx sdk.Context, addr sdk.Address)

Returns the ``sdk.Address`` of the keeper.

::

    keeper.SetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)

Sets the an Address for keeper. Returns ``error`` if fails.
::

    keeper.HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)

Returns a ``BaseAccount`` with Address ``addr``.

::

    keeper.SubtractCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)

Returns the ``sdk.Address`` of the keeper. Returns ``error`` if fails.

::

    keeper.AddCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)

Sets the an Address for keeper.

::

    keeper.SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins)

Sets coins in the keeper. Returns ``error`` if fails.

::

    keeper.InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output)

Gets the corresponding sequence of keeper.


type SendKeeper struct {
	am sdk.AccountMapper
}


type ViewKeeper struct {
	am sdk.AccountMapper
}
