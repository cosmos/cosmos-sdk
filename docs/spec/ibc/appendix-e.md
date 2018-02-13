## Appendix E: Tendermint Header Proofs

TODO: clean this all up

This is a mess now, we need to figure out what formats we use, define go-wire, etc. or just point to the source???? Will do more later, need help here from the tendermint core team.

In order to prove a merkle root, we must fully define the headers, signatures, and validator information returned from the Tendermint consensus engine, as well as the rules by which to verify a header. We also define here the messages used for creating and removing connections to other blockchains as well as how to handle forks.

**Building Blocks: Header, PubKey, Signature, Commit, ValidatorSet**

**-> needs input/support from Tendermint Core team (and go-crypto)**

**Registering Chain**

**Updating Header**

**Validator Changes**

ROOT of trust

As mentioned in the definitions, all proofs are based on an original assumption. The root of trust here is either the genesis block (if it is newer than the unbonding period) or any signed header of the other chain.

When governance on a pair of chain, the respective chains must agree to a root of trust on the counterparty chain. This can be the genesis block on a chain that launches with an IBC channel or a later block header.

From this signed header, one can check the validator set against the validator hash stored in the header, and then verify the signatures match. This provides internal consistency and accountability, but if 5 nodes provide you different headers (eg. of forks), you must make a subjective decision which one to trust. This should be performed by on-chain governance to avoid an exploitable position of trust.

VERIFYING HEADERS

Once we have a trusted header with a known validator set, we can quickly validate any new header with the same validator set. To validate a new header, simply verifying that the validator hash has not changed, and that over 2/3 of the voting power in that set has properly signed a commit for that header. We can skip all intervening headers, as we have complete finality (no forks) and accountability (to punish a double-sign).

This is safe as long as we have a valid signed header by the trusted validator set that is within the unbonding period for staking. In that case, if we were given a false (forked) header, we could use this as proof to slash the stake of all the double-signing validators. This demonstrates the importance of attribution and is the same security guarantee of any non-validating full node. Even in the presence of some ultra-powerful malicious actors, this makes the cost of creating a fake proof for a header equal to at least one third of all staked tokens, which should be significantly higher than any gain of a false message.

UPDATING VALIDATORS SET

If the validator hash is different than the trusted one, we must simultaneously both verify that if the change is valid while, as well as use using the new set to validate the header.  Since the entire validator set is not provided by default when we give a header and commit votes, this must be provided as extra data to the certifier.

A validator change in Tendermint can be securely verified with the following checks:



*   First, that the new header, validators, and signatures are internally consistent
    *   We have a new set of validators that matches the hash on the new header
    *   At least 2/3 of the voting power of the new set validates the new header
*   Second, that the new header is also valid in the eyes of our trust set
    *   Verify at least 2/3 of the voting power of our trusted set, which are also in the new set, properly signed a commit to the new header

In that case, we can update to this header, and update the trusted validator set, with the same guarantees as above (the ability to slash at least one third of all staked tokens on any false proof).


