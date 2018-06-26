Replay Protection
-----------------

In order to prevent `replay
attacks <https://en.wikipedia.org/wiki/Replay_attack>`__ a multi account
nonce system has been constructed as a module, which can be found in
``modules/nonce``. By adding the nonce module to the stack, each
transaction is verified for authenticity against replay attacks. This is
achieved by requiring that a new signed copy of the sequence number
which must be exactly 1 greater than the sequence number of the previous
transaction. A distinct sequence number is assigned per chain-id,
application, and group of signers. Each sequence number is tracked as a
nonce-store entry where the key is the marshaled list of actors after
having been sorted by chain, app, and address.

.. code:: golang

    // Tx - Nonce transaction structure, contains list of signers and current sequence number
    type Tx struct {
        Sequence uint32           `json:"sequence"`
        Signers  []basecoin.Actor `json:"signers"`
        Tx       basecoin.Tx      `json:"tx"`
    }

By distinguishing sequence numbers across groups of Signers,
multi-signature Actors need not lock up use of their Address while
waiting for all the members of a multi-sig transaction to occur. Instead
only the multi-sig account will be locked, while other accounts
belonging to that signer can be used and signed with other sequence
numbers.

By abstracting out the nonce module in the stack, entire series of
transactions can occur without needing to verify the nonce for each
member of the series. An common example is a stack which will send coins
and charge a fee. Within the SDK this can be achieved using separate
modules in a stack, one to send the coins and the other to charge the
fee, however both modules do not need to check the nonce. This can occur
as a separate module earlier in the stack.
