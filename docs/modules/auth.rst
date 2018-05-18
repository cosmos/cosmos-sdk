Auth
====

Defines a standard account structure (``BaseAccount``) and how transaction signers are authenticated.

**BaseAccount**

- **Address** (``sdk.Address``) -
- **Coins** (``sdk.Coins``) -
- **PubKey** (``crypto.PubKey``) -
- **Sequence** (``int64``) -

Methods
^^^^^^^

``NewBaseAccountWithAddress(addr sdk.Address)``

  Returns a ``BaseAccount`` with Address ``addr``.

``baseAccount.GetAddress()``

  Returns the ``sdk.Address`` of the baseAccount.

``baseAccount.SetAddress(addr sdk.Address)``

  Sets the an Address for baseAccount. Returns ``error`` if fails.

``baseAccount.GetPubKey()``

  Returns a ``BaseAccount`` with Address ``addr``.

``baseAccount.SetPubKey(pubKey crypto.PubKey)``

  Returns the ``sdk.Address`` of the baseAccount. Returns ``error`` if fails.

``baseAccount.GetCoins()``

  Sets the an Address for baseAccount.

``baseAccount.setCoins(coins sdk.Coins)``

  Sets coins in the baseAccount. Returns ``error`` if fails.

``baseAccount.GetSequence()``

  Gets the corresponding sequence of baseAccount.

``baseAccount.SetSequence(seq int64)``

  Sets a sequence for baseAccount. Returns ``error`` if fails.

``RegisterBaseAccount(cdc *wire.Codec)``

  Registers BaseAcount in the codec. Useful for testing.

AnteHandler
-----------

``NewAnteHandler(accountMapper sdk.AccountMapper, feeHandler sdk.FeeHandler)``

  Returns an ``AnteHandler`` that checks and increments sequence numbers, checks signatures and deducts fees from the first signer.

``processSig(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address, sig sdk.StdSignature, signBytes []byte)``

  Returns ``sdk.Account`` and ``sdk.Result``. ``processSig`` verifies the signature and increments the sequence. If the account doesn't have a pubkey, set it.

``deductFees(acc sdk.Account, fee sdk.StdFee)``

  Returns ``sdk.Account`` and ``sdk.Result``. Deducts the fee from the account.


Context
-------

``WithSigners(ctx types.Context, accounts []types.Account)``

  Adds the signers in a list of ``Account`` s to the ``Context`` and returns it.

``GetSigners(ctx types.Context)``

  Returns ``sdk.Account`` and ``sdk.Result``. ``processSig`` verifies the signature and increments the sequence. If the account doesn't have a pubkey, set it.

Mapper
------

``AccountMapper`` encodes/decodes accounts using the ``go-amino`` (binary) encoding/decoding library.

**AccountMapper**

- **key** (``sdk.StoreKey``) - The (unexposed) key used to access the store from the Context.
- **proto** (``sdk.Account``) - The prototypical ``sdk.Account`` concrete type.
- **cdc** (``wire.Codec``) - The wire codec for binary encoding/decoding of accounts.

``NewAccountMapper(cdc *wire.Codec, key sdk.StoreKey, proto sdk.Account)``

  Returns a new ``sdk.AccountMapper``.


``am.NewAccountWithAddress(ctx sdk.Context, addr sdk.Address)``

  Returns a new ``sdk.Account`` with a given ``Address``.

``am.GetAccount(ctx sdk.Context, addr sdk.Address)``

  Gets a the ``Account`` struct of a given ``Address``.

``am.SetAccount(ctx sdk.Context, acc sdk.Account)``

  Encodes and saves an ``Account`` in the context's ``KVStore``.

``am.clonePrototype()``

  Creates and returns a ``Account`` struct (or pointer to struct) from ``am.proto``.

``am.encodeAccount(acc sdk.Account)``

  Returns the encoded bytes of an ``Account``.

``am.decodeAccount(bz []byte)``

  Returns a decoded ``Account`` from the enconded account's bytes.
