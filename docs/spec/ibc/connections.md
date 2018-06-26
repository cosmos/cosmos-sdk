## 2 Connections

([Back to table of contents](README.md#contents))

The basis of IBC is the ability to verify in the on-chain consensus ruleset of chain `B` that a data packet received on chain `B` was correctly generated on chain `A`. This establishes a cross-chain linearity guarantee: upon validation of that packet on chain `B` we know that the packet has been executed on chain `A` and any associated logic resolved (such as assets being escrowed), and we can safely perform application logic on chain `B` (such as generating vouchers on chain `B` for the chain `A` assets which can later be redeemed with a packet in the opposite direction).

This section outlines the abstraction of an IBC _connection_: the state and consensus ruleset necessary to perform IBC packet verification. 

### 2.1 Definitions

- Chain `A` is the source blockchain from which the IBC packet is sent
- Chain `B` is the destination blockchain on which the IBC packet is received
- `H_h` is the signed header of chain `A` at height `h`
- `C_h` is a subset of the consensus ruleset of chain `A` at height `h`
- `V_kh` is the value stored on chain `A` under key `k` at height `h`
- `P` is the unbonding period of chain `P`, in units of time
- `dt(a, b)` is the time difference between events `a` and `b`

Note that of all these, only `H_h` defines a signature and is thus attributable.

### 2.2 Requirements

To facilitate an IBC connection, the two blockchains must provide the following proofs:

1. Given a trusted `H_h` and `C_h` and an attributable update message `U_h`,  
   it is possible to prove `H_h'` where `C_h' == C_h` and `dt(now, H_h) < P`
2. Given a trusted `H_h` and `C_h` and an attributable change message `X_h`,  
   it is possible to prove `H_h'` where `C_h' /= C_h` and `dt(now, H_h) < P`
3. Given a trusted `H_h` and a Merkle proof `M_kvh` it is possible to prove `V_kh`

It is possible to make use of the structure of BFT consensus to construct extremely lightweight and provable messages `U_h'` and `X_h'`. The implementation of these requirements with Tendermint consensus is defined in [Appendix E](appendices.md#appendix-e-tendermint-header-proofs). Another algorithm able to provide equally strong guarantees (such as Casper) is also compatible with IBC but must define its own set of update and change messages.

The Merkle proof `M_kvh` is a well-defined concept in the blockchain space, and provides a compact proof that the key value pair `(k, v)` is consistent with a Merkle root stored in `H_h`. Handling the case where `k` is not in the store requires a separate proof of non-existence, which is not supported by all Merkle stores. Thus, we define the proof only as a proof of existence. There is no valid proof for missing keys, and we design the algorithm to work without it.

Blockchains supporting IBC must implement Merkle proof verification:

`valid(H_h, M_kvh) ⇒ true | false`

### 2.3 Connection Lifecycle

#### 2.3.1 Opening a connection

All proofs require an initial `H_h` and `C_h` for some `h`, where `dt(now, H_h) < P`.

Establishing a bidirectional initial root-of-trust between the two blockchains (`A` to `B` and `B` to `A`) — `H_ah` and `C_ah` stored on chain `B`, and `H_bh` and `C_bh` stored on chain `A` — is necessary before any IBC packets can be sent. 

Any header may be from a malicious chain (e.g. shadowing a real chain state with a fake validator set), so a subjective decision is required before establishing a connection. This can be performed permissionlessly, in which case users later utilizing the IBC channel must check the root-of-trust themselves, or authorized by on-chain governance for additional assurance.

#### 2.3.2 Following block headers

We define two messages `U_h` and `X_h`, which together allow us to securely advance our trust from some known `H_n` to some future `H_h` where `h > n`. Some implementations may require that `h == n + 1` (all headers must be processed in order). IBC implemented on top of Tendermint or similar BFT algorithms requires only that `delta-vals(C_n, C_h) < ⅓` (each step must have a change of less than one-third of the validator set)[[4](./references.md#4)].

Either requirement is compatible with IBC. However, by supporting proofs where  `h - n > 1`, we can follow the block headers much more efficiently in situations where the majority of blocks do not include an IBC packet between chains `A` and `B`, and enable low-bandwidth connections to be implemented at very low cost. If there are packets to relay every block, these two requirements collapse to the same case (every header must be relayed).

Since these messages `U_h` and `X_h` provide all knowledge of the remote blockchain, we require that they not just be provable, but also attributable. As such, any attempt to violate the finality guarantees in headers posted to chain `B` can be submitted back to chain `A` for punishment, in the same manner that chain `A` would independently punish (slash) identified Byzantine actors.

More formally, given existing set of trust `T` =  `{(H_i, C_i), (H_j, C_j), …}`, we must provide:

`valid(T, X_h | U_h) ⇒ true | false | unknown`

`valid` must fulfill the following properties:

```
if H_h-1 ∈ T then
  valid(T, X_h | U_h) ⇒ true | false
  ∃ (U_h | X_h) ⇒ valid(T, X_h | U_h)
```

```
if C_h ∉ T then
  valid(T, U_h) ⇒ false
```

We can then process update transactions as follows:

`update(T, X_h | U_h) ⇒ success | failure`

```
update(T, X_h | U_h) = match valid(T, X_h | U_h) with
  false ⇒ fail with "invalid proof"
  unknown ⇒ fail with "need a proof between current and h"
  true ⇒ 
    set T = T ∪ (H_h, C_h)
```

Define `max(T)` as `max(h, where H_h ∈ T)`. For any `T` with `max(T) == h-1`, there must exist some `X_h | U_h` so that `max(update(T, X_h | U_h)) == h`.
By induction, there must exist a set of proofs, such that `max(update…(T,...)) == h + n` for any `n`.

Bisection can be used to discover this set of proofs. That is, given `max(T) == n` and `valid(T, X_h | U_h) == unknown`, we then try `update(T, X_b | U_b)`, where _`b == (h + n) / 2`. The base case is where `valid(T, X_h | U_h) == true` and is guaranteed to exist if `h == max(T) + 1`.

#### 2.3.3 Closing a connection

IBC implementations may optionally include the ability to close an IBC connection and prevent further header updates, simply causing `update(T, X_h | U_h)` as defined above to always return `false`.

Closing a connection may break application invariants (such as fungiblity - token vouchers on chain `B` will no longer be redeemable for tokens on chain `A`) and should only be undertaken in extreme circumstances such as Byzantine behavior of the connected chain.

Closure may be permissioned to an on-chain governance system, an identifiable party on the other chain (such as a signer quorum, although this will not work in some Byzantine cases), or any user who submits an application-specific fraud proof. When a connection is closed, application-specific measures may be undertaken to recover assets held on a Byzantine chain. We defer further discussion to [Appendix D](appendices.md#appendix-d-byzantine-recovery-strategies).
