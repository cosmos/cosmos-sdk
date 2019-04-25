# Cosmos Hub to Launch Mainnet

## Pre-launch Dependencies & How to Safely Claim Your Atoms

In the summer of 2016, the [Cosmos whitepaper][whitepaper] was released. In the
spring of 2017, the [Cosmos fundraiser][fundraiser] was completed. In the first
months of 2019, the software is [feature complete][releases]. The launch of the
Cosmos Hub draws near. What does this mean for Atom holders?

If you are an Atom holder, you will be able to delegate Atoms to validators on
the main network and vote on governance proposals. In fact, the future success
of the network depends on you responsibly doing so! However, you will not be
able to transfer Atoms yet. Transfers will be disabled at the protocol level
until a hard-fork is executed to enable them.

Atom holders should carefully follow the guidelines in order to safely delegate
Atoms. Please read through the entire guide first to familiarize yourself
before you actually do anything: [CLI guide][cli]

The process outlined in the guide is currently the only verified and secure way
to delegate Atoms at launch. This is because the gaiacli tool used in the guide
is the only wallet software undergoing third-party security audits right now.
No other wallet providers have begun security audits yet.

Remember that delegating Atoms involves significant risk. Once delegated to a
validator, Atoms are bonded for a period of time during which they cannot be
recovered. If the validator misbehaves during this time, some or all of the
delegated Atoms may be burned. It is your responsibility to perform due
diligence on validators before delegating!

The Cosmos Hub is highly experimental software. In these early days, we can
expect to have issues, updates, and bugs. The existing tools require advanced
technical skills and involve risks which are outside of the control of the
Interchain Foundation and/or the Tendermint team (see also the risk section in
the [Interchain Cosmos Contribution Terms][terms]). Any use of this open source
[Apache 2.0 licensed][apache] software is done at your own risk and on a “AS
IS” basis without warranties or conditions of any kind, and any and all
liability of the Interchain Foundation and/or the Tendermint team for damages
arising in connection to the software is excluded. Please exercise extreme
caution!

If you are looking for more information about delegation and want to talk to
the folks developing Cosmos, join the virtual meetup on February 14 where you
will be walked through the step-by-step instructions for delegating Atoms at
launch.

Register here: [gotowebinar.com/register/][webinar]

## Remaining Milestones for Launch

To follow mainnet launch progress, please bookmark:
[cosmos.network/launch][cosmos].

### 5 Cosmos-SDK Security Audits ✔

In early January, the Cosmos-SDK underwent the first in a series of third-party
security assessments scheduled for Q1 2019. This audit took place over a two
and a half week period. To date, two different security auditing firms have
assessed various parts of the Cosmos-SDK and a third audit is under way.

### 4 Cosmos SDK Feature Freeze

The final breaking changes to the Cosmos-SDK are included in the [v0.31.0
launch RC][rc]. Once this RC is completed, the Cosmos-SDK team will engage in a
round of internal bug hunting to further ensure sufficient pre-launch security
due diligence.

Right after Cosmos-SDK v0.31.0 is released, a Gaia testnet will be released in
an effort to flush out any hard to find bugs.

### 3 Game of Stakes Completed

Game of Stakes (GoS), [the first adversarial testnet competition of its
kind][gos], was launched in December 2018 to stress test the economic incentive
and social layers of a blockchain network secured purely by Proof-of-Stake. The
GoS blockchain was successfully hard-forked three times to date. As soon as the
GoS concludes, the [scoring criteria][scoring] will be used to determine
winners. Those will be announced following the completion of the game.

### 2 Genesis Transactions Collected

The Interchain Foundation will publish a recommendation for the allocation of
Atoms at genesis. This will include allocations for Cosmos fundraiser
participants, early contributors, and Game of Stakes winners. Any one with a
recommended allocation will have the opportunity to submit a gentx, which is
required to become a validator at genesis. The ultimate result of the
recommended allocation and the collection of gentxs is a final [genesis
file][file].

### 1 Cosmos Hub Mainnet Launch

Once a genesis file is adopted by the community, and +⅔ of the voting power
comes online, the Cosmos mainnet will be live.

## Official Cosmos Communication Channels

These are the official accounts that will communicate launch details:

- [Cosmos Network](https://twitter.com/cosmos)
- [Cosmos GitHub](https://github.com/cosmos)
- [Cosmos Blog](https://blog.cosmos.network)

Please be aware that the [Cosmos forum][forum], [Riot chat groups][riot], and
[Telegram group][telegram] should not be treated as official news from Cosmos.

If you have doubt or confusion about what next steps to take and are unsure
about trustworthy sources of information, do nothing for the initial period and
wait for an update via the three communication channels listed above. Do not
ever provide your 12 words to any admin, websites or unofficial software.

**We will never ask you for your private key or your seed phrase.**

## Staying Safe (and Secure!) for Mainnet Launch

The launch of any public blockchain is an incredibly exciting time, and it’s
definitely one that malicious actors may try to take advantage of for their own
personal gain. [Social engineering][social] has existed for about as long as
human beings have been on the planet, and in the technical era, it usually
takes in the form of [phishing] or [spearphishing]. Both of these attacks are
wildly successful forms of trickery that are responsible for over 95% of
account security breaches, and they don’t just happen via email: these days,
opportunistic and targeted phishing attempts take place [anywhere that you have
an inbox][inbox]. It doesn’t matter if you’re using Signal, Telegram, SMS,
Twitter, or just checking your DMs on forums or social networks, attackers have
a [plethora of opportunities][opportunities] to gain foothold in your digital
life in effort to separate you from valuable information and assets that you
most definitely don’t want to lose.

While the prospect of having to deal with a malicious actor plotting against
you may seem daunting, there are many things that you can do to protect
yourself from all kinds of social engineering schemes. In terms of preparing
for mainnet launch, this should require training your instincts to successfully
detect and avoid security risks, curating resources to serve as a source of
truth for verifying information, and going through a few technical steps to
reduce or eliminate the risk of key or credential theft.

**Here are few rules of engagement to keep in mind when you’re preparing for
Cosmos mainnet launch:**

- Download software directly from official sources, and make sure that you’re
  always using the latest, most secure version of gaiacli when you’re doing
  anything that involves your 12 words. The latest versions of Tendermint, the
  Cosmos-SDK, and gaiacli will always be available from our official GitHub
  repositories, and downloading them from there ensures that you will not be
  tricked into using a maliciously modified version of software.

- Do not share your 12 words with anyone. The only person who should ever need
  to know them is you. This is especially important if you’re ever approached
  by someone attempting to offer custodial services for your Atom: to avoid
  losing control of your tokens, you should store them offline to minimize the
  risk of theft and have a strong backup strategy in place. And never, ever
  share them with anyone else.

- Be skeptical of unexpected attachments or emails that ask you to visit a
  suspicious or unfamiliar website in the context of blockchains or
  cryptocurrency. An attacker may attempt to lure you to a [compromised site]
  designed to steal sensitive information from your computer. If you’re a Gmail
  user, test your resilience against the latest email-based phishing tactics
  [here][quiz].

- Do your due diligence before purchasing Atoms. Atoms will not be transferable
  at launch, so they *cannot* be bought or sold until a hard fork enables them
  to be. If and when they become transferable, make sure that you’ve researched
  the seller or exchange to confirm that the Atoms are coming from a
  trustworthy source.

- Neither the Tendermint team nor the Interchain Foundation will be selling
  Atoms, so if you see social media posts or emails advertising a token sale
  from us, they’re not real and should be avoided.  Enable 2-factor
  authentication, and be mindful of recovery methods used to regain access to
  your most important accounts. Unprotected accounts like email, social media,
  your GitHub account, the Cosmos Forum and anything in between could give an
  attacker opportunities to gain foothold in your online life. If you haven’t
  done so yet, start using an authenticator app or a hardware key immediately
  wherever you manage your tokens. This is a simple, effective, and proven way
  to reduce the risk of account theft.

- Be skeptical of technical advice, especially advice that comes from people
  you do not know in forums and on group chat channels. Familiarize yourself
  with important commands, especially those that will help you carry out
  high-risk actions, and consult our official documentation to make sure that
  you’re not being tricked into doing something that will harm you or your
  validator. And remember that the Cosmos forum, Riot channels, and Telegram
  are not sources of official information or news about Cosmos.

- Verify transactions before hitting send. Yes, those address strings are long,
  but visually comparing them in blocks of 4 characters at a time may be the
  difference between sending them to the right place or sending them into
  oblivion.

*If a deal pops up that [sounds too good to be true][good], or a message shows
up asking for information that should never, ever be shared with someone else,
you can always work to verify it before engaging with it by navigating to a
website or official Cosmos communication channel on your own. No one from
Cosmos, the Tendermint team or the Interchain Foundation will ever send an
email that asks for you to share any kind of account credentials or your 12
words with us, and we will always use our official blog, Twitter and GitHub
accounts to communicate important news directly to the Cosmos community.*

[whitepaper]: https://cosmos.network/resources/whitepaper
[fundraiser]: https://fundraiser.cosmos.network/
[releases]: https://github.com/cosmos/cosmos-sdk/releases
[cosmos]: https://cosmos.network/launch
[social]: https://en.wikipedia.org/wiki/Social_engineering_%28security%29
[phishing]: https://ssd.eff.org/en/module/how-avoid-phishing-attacks
[spearphishing]: https://en.wikipedia.org/wiki/Phishing#Spear_phishing
[inbox]: https://www.umass.edu/it/security/phishing-fraudulent-emails-text-messages-phone-calls
[opportunities]: https://jia.sipa.columbia.edu/weaponization-social-media-spear-phishing-and-cyberattacks-democracy
[cli]: https://github.com/cosmos/cosmos-sdk/blob/develop/docs/gaia/delegator-guide-cli.md
[webinar]: https://register.gotowebinar.com/register/5028753165739687691
[terms]: https://github.com/cosmos/cosmos/blob/master/fundraiser/Interchain%20Cosmos%20Contribution%20Terms%20-%20FINAL.pdf
[apache]: https://www.apache.org/licenses/LICENSE-2.0
[gos]: https://blog.cosmos.network/announcing-incentivized-testnet-game-efe64e0956f6
[scoring]: https://github.com/cosmos/game-of-stakes/blob/master/README.md#scoring
[file]: https://forum.cosmos.network/t/genesis-files-network-starts-vs-upgrades/1464
[forum]: https://forum.cosmos.network/
[riot]: https://riot.im/app/#/group/+cosmos:matrix.org
[telegram]: http://t.me/cosmosproject
[good]: https://www.psychologytoday.com/us/blog/mind-in-the-machine/201712/how-fear-is-being-used-manipulate-cryptocurrency-markets
[rc]: https://github.com/cosmos/cosmos-sdk/projects/27
[compromised site]: https://blog.malwarebytes.com/cybercrime/2013/02/tools-of-the-trade-exploit-kits/
[quiz]: https://phishingquiz.withgoogle.com/
