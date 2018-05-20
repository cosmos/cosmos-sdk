Bank
====

The bank module defines how coins (*i.e.* cryptocurrencies) are transferred.

Messages
--------

**MsgSend**
^^^^^^^^^^^

- **Inputs** (``[]Input``) -
- **Outputs** (``[]Output``) -

The ``Input`` and ``Output`` structs define an ``Address`` and a set of ``Coins``.

Methods
"""""""

``NewMsgSend(in []Input, out []Output)``
****************************************

  Returns: ``MsgSend``

  Creates a message to send coins.

``msg.Type()``
**************

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message. Returns error if fails.

**MsgIssue**
^^^^^^^^^^^^

- **Banker** (``sdk.Address``) -
- **Outputs** (``]Output``) -

Methods
"""""""

``NewMsgIssue(banker sdk.Address, out []Output)``
**************************************************

  Returns: ``MsgIssue``

  Creates a message to issue coins.

``msg.Type()``
**************

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message. Returns error if fails.

Keepers
-------

A Bank ``Keeper`` is basically an ``AccountMapper`` that manages transfers between accounts.

There are 3 types of bank keepers:

- ``ViewKeeper``: only allows reading of balances
- ``SendKeeper``: only allows transfers between accounts, without the possibility of creating coins
- ``Keeper``: all the above, plus allowing the creation and deletion of coins to an ``Address``

**Keeper/ViewKeeper/SendKeeper**
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- **am** (``sdk.AccountMapper``) - Keeper account mapper

Methods
"""""""

``NewKeeper(am sdk.AccountMapper)``
***********************************

  Returns: ``Keeper``

  Creates a new keeper.

``keeper.GetCoins(ctx sdk.Context, addr sdk.Address)``
******************************************************

  Returns: ``sdk.Coins``

  Gets the coins at the given address.

``keeper.SetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``
*********************************************************************

  Returns: ``sdk.Error``

  Sets the coins at the address.

``keeper.HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``
*********************************************************************

  Returns: ``bool``

  Checks whether or not an account has at least a given amount of coins.

``keeper.SubtractCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``
**************************************************************************

  Returns: ``sdk.Coins``, ``sdk.Tags``, ``sdk.Error``

  Subtracts a given amount of coins held by an address. Returns error if fails.

``keeper.AddCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins)``
*********************************************************************

  Returns: ``sdk.Coins``, ``sdk.Tags``, ``sdk.Error``

  Adds a given amount of coins held by an address. Returns error if fails.


``keeper.SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins)``
**********************************************************************************************

  Returns: ``sdk.Coins``, ``sdk.Tags``, ``sdk.Error``

  Moves coins from one account to another. Returns error if fails.

``keeper.InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output)``
******************************************************************************

  Returns: ``sdk.Coins``, ``sdk.Tags``, ``sdk.Error``

  Handles a list of inputs and outputs.  Returns error if fails.

Handlers
--------

Bank Handlers
^^^^^^^^^^^^^

Methods
"""""""

``NewHandler(k Keeper)``
************************

  Returns: ``sdk.Handler``

  Creates a handler for bank type messages.

``handleMsgSend(ctx sdk.Context, k Keeper, msg MsgSend)``
*********************************************************

  Returns: ``sdk.Result``

  Handles the logic behind sending coins.
