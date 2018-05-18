Bank
====

Defines how coins (i.e cryptocurrencies) are transferred.

**MsgSend**

- **Inputs** (``[]Input``) -
- **Outputs** (``[]Output``) -


The Input and Output structs are define an ``Address`` and a set of ``Coins``.

**Input**

- **Address** (``sdk.Address``) -
- **Coins** (``sdk.Coins``) -

**Output**

- **Address** (``sdk.Address``) -
- **Coins** (``sdk.Coins``) -

**MsgIssue**

- **Banker** (``sdk.Address``) -
- **Outputs** (``]Output``) -

``NewMsgSend(in []Input, out []Output)``

  Returns a new ``MsgSend``.

``NewMsgIssue(banker sdk.Address, out []Output)``
  Returns a new ``MsgIssue``.

``msg.SetAddress(addr sdk.Address)``

  Sets the an Address for baseAccount. Returns ``error`` if fails.

``msg.ValidateBasic()``

  Validates the correctness of a ``MsgIssue`` attributes. Returns an ``sdk.Error`` if fails.

``msg.GetSignBytes()``

  Returns the signatures' ``bytes`` of the ``MsgIssue``.

``msg.GetSigners()``

  Returns the addresses of the signers of the ``MsgIssue``.

Keeper
------

A Bank ``Keeper`` is basically an ``AccountMapper`` that manages transfers between accounts.

There are 3 types of bank keepers:

- ``ViewKeeper``: only allows reading of balances
- ``SendKeeper``: only allows transfers between accounts, without the possibility of creating coins
- ``Keeper``: all the above, plus allowing the creation and deletion of coins to an ``Address``

**Keeper/ViewKeeper/SendKeeper**

- **am** (``sdk.AccountMapper``) - Keeper account mapper

Methods

``NewKeeper(am sdk.AccountMapper)``

  Returns a new ``Keeper``.

``keeper.GetCoins(ctx sdk.Context, addr sdk.Address)``

  Returns the ``sdk.Address`` of the keeper.

``keeper.SetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``

  Sets the an Address for keeper. Returns ``error`` if fails.

``keeper.HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``

  Returns a ``BaseAccount`` with Address ``addr``.

``keeper.SubtractCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``

  Returns the ``sdk.Address`` of the keeper. Returns ``error`` if fails.

``keeper.AddCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``

  Sets the an Address for keeper.

``keeper.SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins)``

  Sets coins in the keeper. Returns ``error`` if fails.

``keeper.InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output)``

  Gets the corresponding sequence of keeper.
