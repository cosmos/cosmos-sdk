# Security

> **IMPORTANT**: If you find a security issue, you can contact our team directly at
security@tendermint.com, or report it to our [bug bounty program](https://hackerone.com/tendermint) on HackerOne. *DO NOT* open a public issue on the repository.

## Bug Bounty

As part of our [Coordinated Vulnerability Disclosure Policy](https://tendermint.com/security), we operate a
[bug bounty program](https://hackerone.com/tendermint) with Hacker One.

See the policy linked above for more details on submissions and rewards and read
this [blog post](https://blog.cosmos.network/bug-bounty-program-for-tendermint-cosmos-833c67693586) for the program scope.

The following is a list of examples of the kinds of bugs we're most interested
in for the Cosmos SDK. See [here](https://github.com/tendermint/tendermint/blob/master/SECURITY.md) for vulnerabilities we are interested
in for Tendermint and other lower-level libraries (eg. [IAVL](https://github.com/tendermint/iavl)).

### Core packages

- [`/baseapp`](https://github.com/cosmos/cosmos-sdk/tree/master/baseapp)
- [`/crypto`](https://github.com/cosmos/cosmos-sdk/tree/master/crypto)
- [`/types`](https://github.com/cosmos/cosmos-sdk/tree/master/types)
- [`/store`](https://github.com/cosmos/cosmos-sdk/tree/master/store)

### Modules

- [`x/auth`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth)
- [`x/bank`](https://github.com/cosmos/cosmos-sdk/tree/master/x/bank)
- [`x/capability`](https://github.com/cosmos/cosmos-sdk/tree/master/x/capability)
- [`x/staking`](https://github.com/cosmos/cosmos-sdk/tree/master/x/staking)
- [`x/slashing`](https://github.com/cosmos/cosmos-sdk/tree/master/x/slashing)
- [`x/evidence`](https://github.com/cosmos/cosmos-sdk/tree/master/x/evidence)
- [`x/distribution`](https://github.com/cosmos/cosmos-sdk/tree/master/x/distribution)
- [`x/ibc`](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc)
- [`x/ibc-transfer`](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc-transfer)
- [`x/mint`](https://github.com/cosmos/cosmos-sdk/tree/master/x/mint)

We are interested in bugs in other modules, however the above are most likely to
have significant vulnerabilities, due to the complexity / nuance involved. We
also recommend you to read the [specification](https://github.com/cosmos/cosmos-sdk/blob/master/docs/building-modules/README.md) of each module before digging into
the code.

### How we process Tx parameters

- Integer operations on tx parameters, especially `sdk.Int` / `sdk.Dec`
- Gas calculation & parameter choices
- Tx signature verification (see [`x/auth/ante`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth/ante))
- Possible Node DoS vectors (perhaps due to gas weighting / non constant timing)

### Handling private keys

- HD key derivation, local and Ledger, and all key-management functionality
- Side-channel attack vectors with our implementations
  - e.g. key exfiltration based on time or memory-access patterns when decrypting privkey
  
## Disclosure Process

The Cosmos SDK team uses the following disclosure process:

1. Once a security report is received, the Cosmos SDK team works to verify the issue and confirm its severity level using CVSS.
1. The Cosmos SDK team collaborates with the Tendermint and Gaia teams to determine the vulnerability’s potential impact on the Cosmos Hub and partners.
1. Patches are prepared for eligible releases of Cosmos SDK in private repositories. See “Supported Releases” below for more information on which releases are considered eligible.
1. If it is determined that a CVE-ID is required, we request a CVE through a CVE Numbering Authority.
1. We notify the community that a security release is coming, to give users time to prepare their systems for the update. Notifications can include forum posts, tweets, and emails to partners and validators.
1. 24 hours following this notification, the fixes are applied publicly and new releases are issued.
1. Gaia updates their Tendermint Core and Cosmos SDK dependencies to use these releases, and then themselves issue new releases.
1. Once releases are available for Tendermint Core, Cosmos SDK and Gaia, we notify the community, again, through the same channels as above. We also publish a Security Advisory on Github and publish the CVE, as long as neither the Security Advisory nor the CVE include any information on how to exploit these vulnerabilities beyond what information is already available in the patch itself.
1. Once the community is notified, we will pay out any relevant bug bounties to submitters.
1. One week after the releases go out, we will publish a post with further details on the vulnerability as well as our response to it.

This process can take some time. Every effort will be made to handle the bug in as timely a manner as possible, however it's important that we follow the process described above to ensure that disclosures are handled consistently and to keep Cosmos SDK and its downstream dependent projects--including but not limited to Gaia and the Cosmos Hub--as secure as possible.

### Disclosure Communications

Communications to partners will usually include the following details:
1. Affected version(s)
1. New release version
1. Impact on user funds
1. For timed releases, a date and time that the new release will be made available
1. Impact on the partners if upgrades are not completed in a timely manner
1. Potential actions to take if an adverse condition arises during the security release process

An example notice looks like:
```
Dear Cosmos SDK partners,

A critical security vulnerability has been identified in Cosmos SDK vX.X.X. 
User funds are NOT at risk; however, the vulnerability can result in a chain halt.

This notice is to inform you that on [[**March 1 at 1pm EST/6pm UTC**]], we will be releasing Cosmos SDK vX.X.Y, which patches the security issue. 
We ask all validators to upgrade their nodes ASAP.

If the chain halts, validators with sufficient voting power need to upgrade and come online in order for the chain to resume.
```

### Example Timeline

The following is an example timeline for the triage and response. The required roles and team members are described in parentheses after each task; however, multiple people can play each role and each person may play multiple roles.

#### 24+ Hours Before Release Time

1. Request CVE number (ADMIN)
1. Gather emails and other contact info for validators (COMMS LEAD)
1. Test fixes on a testnet  (COSMOS SDK ENG)
1. Write “Security Advisory” for forum (COSMOS SDK LEAD)

#### 24 Hours Before Release Time

1. Post “Security Advisory” pre-notification on forum (COSMOS SDK LEAD)
1. Post Tweet linking to forum post (COMMS LEAD)
1. Announce security advisory/link to post in various other social channels (Telegram, Discord) (COMMS LEAD)
1. Send emails to partners or other users (PARTNERSHIPS LEAD)

#### Release Time

1. Cut Cosmos SDK releases for eligible versions (COSMOS SDK ENG)
1. Cut Gaia release for eligible versions (GAIA ENG)
1. Post “Security releases” on forum (COSMOS SDK LEAD)
1. Post new Tweet linking to forum post (COMMS LEAD)
1. Remind everyone via social channels (Telegram, Discord)  that the release is out (COMMS LEAD)
1. Send emails to validators or other users (COMMS LEAD)
1. Publish Security Advisory and CVE, if CVE has no sensitive information (ADMIN)

#### After Release Time

1. Write forum post with exploit details (COSMOS SDK LEAD)
1. Approve pay-out on HackerOne for submitter (ADMIN)

#### 7 Days After Release Time

1. Publish CVE if it has not yet been published (ADMIN)
1. Publish forum post with exploit details (COSMOS SDK ENG, COSMOS SDK LEAD)
