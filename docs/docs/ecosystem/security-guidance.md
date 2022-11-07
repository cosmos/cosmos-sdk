

**Responding to security incidents on Cosmos blockchains**

**Preamble**

There are no experts here.  Because cosmos chains are socially governed on both the software and infrastructure levels, It is important for the reader to understand that they may need to make guesses.

In general, disclosure of issues is inappropriate, but due to the social nature of these chains, it may at times be necessary to make disclosures and this document is aimed at guiding the reader toward making the best possible decisions under difficult circumstances.

**Safety over liveness**

When one third of the vote power is removed from a Cosmos blockchain, it stops producing blocks, and this eliminates any risk of an attack on the current version of the software.

This technique has saved human lives in disaster scenarios (eg: suicide).  

**Guidelines for chain teams**



* Lint and format your code following the cosmos SDK repository standards
* Write tests, and make sure that your code works how you want it to
* Engage your validators in auditing code and cultivate a culture where validators know the code base.
* Know your validators
* Don't blindly trust upstream (eg: tendermint, iavl, cosmos-sdk, IBC-go): code is made by people, people can make mistakes.  
* If you see a security issue upstream and you think that it may affect multiple chains, contact [security@interchain.io](mailto:security@interchain.io) or a known member of the cosmos SDK, tendermint, or IBC-go developer community. 
* Use a continuous integration system on your repository to ensure that each pull request into your release branch works the way you think it does.
* Treat validators as a key part of your team and security practices.

If there is a known exploitable issue in your code, or your chain is being exploited, the best possible thing that you can do is halt your chain. In order to do that, in your validators discord group, or other coordination group on some kind of chat platform, you need 1/3 of your chains vote power to go offline.  

Being offline is important, because it allows you to think and reason about the situation, without the risk of attack.

Especially if an issue has already been disclosed in public, get the message out to all of your validators as rapidly as possible to take the chain down.

**For validators**



* Ultimately you choose whether or not the chain makes blocks
* If you believe that the chain is secure, make blocks
* If you do not believe that the chain is secure, talk privately to the team or to small groups of trusted validators.
* Validate chains where you feel like you can trust the team and you feel like you are a part of the team because there should not be a distinction between team and validator set.
* Respect the privacy of validator chats and minimize outside discussions of security situations.  This allows a culture where validators can freely discuss and contribute to the chain.
* Have a pre-flight checklist for new chains that you validate where you:
    * Run tests
    * Check that ecosystem modules are the latest released patch version in go.mod
    * Have a look at app.go
    * Check that golangci-lint matches the cosmos-sdk
    * Have a look in the CMD folder for things that look like they don't belong 
    * Speak to the developers
    * Look over the continuous integration system, like GitHub actions, gitlab CI, or circleCI
* Freely disagree with chain teams on security matters, as your actual incentivization should come from your delegators, and you should protect them
* Develop your own networks with people building in the ecosystem, bearing in mind that we are decentralized and will not always agree.  
* Make patches to chains if you feel properly incentivized to do so and that the liability incurred by helping is not greater than the potential incentives.

Validators have a very important role in securing Cosmos.

Use your best judgment and attempt to be fact-driven at all times. The situations that you will face are not always clear, and it is possible to do more harm than good, or the exact inverse.

There is no course to study to get on this career path, however making earnest contributions to chains in the ecosystem is a very good way to ensure your success and to ensure that you follow secure practices.

The economic success of the chains that you validate is related to your ability to validate them in a secure fashion, because that is much more expensive than just signing blocks.

On the chains that you validate, you should be lobbying for clarity when it comes to your role as a validator.  Since Cosmos is decentralized, there really is no one way to do it, and you will experience many different techniques.
