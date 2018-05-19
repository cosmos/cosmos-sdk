Auth
====

Defines a standard account structure (``BaseAccount``) and how transaction signers are authenticated.

BaseAccount
-----------

**BaseAccount**
^^^^^^^^^^^^^^^

- **Address** (``sdk.Address``) -
- **Coins** (``sdk.Coins``) -
- **PubKey** (``crypto.PubKey``) -
- **Sequence** (``int64``) -

Methods
"""""""

``NewBaseAccountWithAddress(addr sdk.Address)``
***********************************************

  Returns: ``BaseAccount``

  Creates a ``BaseAccount`` with Address ``addr``.

``baseAccount.GetAddress()``
****************************

  Returns: ``sdk.Address``

  Returns the associated address of the baseAccount.

``baseAccount.SetAddress(addr sdk.Address)``
********************************************

  Returns: ``error``
  Sets an address for baseAccount. Returns an error if fails.

``baseAccount.GetPubKey()``
***************************

  Returns: ``crypto.PubKey``

  Returns the baseAccount public key.

``baseAccount.SetPubKey(pubKey crypto.PubKey)``
***********************************************

  Returns: ``error``

  Sets a public key for the baseAccount.

``baseAccount.GetCoins()``
**************************

  Returns: ``sdk.Coins``

  Returns an array of Coins held by the baseAccount.

``baseAccount.setCoins(coins sdk.Coins)``
*****************************************

  Returns: ``error``

  Sets coins in the baseAccount. Returns an error if fails.

``baseAccount.GetSequence()``
*****************************

  Returns: ``int64``

  Gets the corresponding sequence of baseAccount.

``baseAccount.SetSequence(seq int64)``
**************************************

  Returns: ``error``

  Sets a sequence for baseAccount. Returns an error if fails.

``RegisterBaseAccount(cdc *wire.Codec)``
****************************************

  Returns: ``nil``

  Registers ``BaseAcount`` in the codec. Useful for testing.

Handlers
--------

AnteHandler
^^^^^^^^^^^

An ``AnteHandler`` is a ``Handler`` that checks and increments sequence numbers, checks signatures and deducts fees from the first signer.

Methods
"""""""

``NewAnteHandler(accountMapper sdk.AccountMapper, feeHandler sdk.FeeHandler)``
******************************************************************************

  Returns: ``AnteHandler``

  Creates a new AnteHandler.

``processSig(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address, sig sdk.StdSignature, signBytes []byte)``
***************************************************************************************************************

  Returns: ``sdk.Account``, ``sdk.Result``

  Verifies the signature and increments the sequence. If the account doesn't have a pubkey, ``processSig`` sets it.

``deductFees(acc sdk.Account, fee sdk.StdFee)``
***********************************************

  Returns: ``sdk.Account``, ``sdk.Result``

  Deducts the fee from the account.


Context
-------

Context
^^^^^^^

Methods
"""""""

``WithSigners(ctx types.Context, accounts []types.Account)``
************************************************************

  Returns: ``types.Context``

  Adds the signers in a list of ``Account`` s to the ``Context`` and returns it.

``GetSigners(ctx types.Context)``
*********************************

  Returns: ``[]types.Account``

  Gets the signers from the Context

Mapper
------

**AccountMapper**
^^^^^^^^^^^^^^^^^

- **key** (``sdk.StoreKey``) - The (unexposed) key used to access the store from the Context.
- **proto** (``sdk.Account``) - The prototypical ``sdk.Account`` concrete type.
- **cdc** (``wire.Codec``) - The wire codec for binary encoding/decoding of accounts.

``AccountMapper`` encodes/decodes accounts using the ``go-amino`` (binary) encoding/decoding library.

Methods
"""""""

``NewAccountMapper(cdc *wire.Codec, key sdk.StoreKey, proto sdk.Account)``
**************************************************************************

  Returns: ``AccountMapper``

  Creates a new ``sdk.AccountMapper``.

``am.NewAccountWithAddress(ctx sdk.Context, addr sdk.Address)``
***************************************************************

  Returns: ``sdk.Account``

  Sets a given ``Address`` to the accountMapper.

``am.GetAccount(ctx sdk.Context, addr sdk.Address)``
****************************************************

  Returns: ``sdk.Account``

  Gets a the associated account of a given address.

``am.SetAccount(ctx sdk.Context, acc sdk.Account)``
***************************************************

  Returns: ``nil``

  Encodes and saves an account in the context's ``KVStore``.

``am.IterateAccounts(ctx sdk.Context, process func(sdk.Account)``
*****************************************************************

  Returns: ``bool``

  Iterates over a context's ``KVStore`` to validate accounts.

``am.GetPubKey(ctx sdk.Context, addr sdk.Address)``
***************************************************

  Returns: ``crypto.PubKey``, ``sdk.Error``

  Returns the *PubKey* of the account at address.

``am.setPubKey(ctx sdk.Context, addr sdk.Address, newPubKey crypto.PubKey)``
****************************************************************************

  Returns: ``sdk.Error``

  Sets the *PubKey* of the account at address.

``am.GetSequence(ctx sdk.Context, addr sdk.Address)``
*****************************************************

  Returns: ``int64``, ``sdk.Error``

  Returns the sequence of the account of a corresponding address.

``am.setSequence(ctx sdk.Context, addr sdk.Address, newSequence int64)``
************************************************************************

  Returns: ``sdk.Error``

  Sets the sequence of the account with the given address.

``am.clonePrototype()``
***********************

  Returns: ``sdk.Account``

  Creates a new ``Account`` struct (or pointer to struct) from ``am.proto``.

``am.encodeAccount(acc sdk.Account)``
*************************************

  Returns: ``[]byte``

  Returns the encoded bytes of an account.

``am.decodeAccount(bz []byte)``
*******************************

  Returns: ``sdk.Account``

  Returns a decoded account from its enconded byte array.
