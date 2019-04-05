# Concepts

## The Minting Mechanism

The minting mechanism was designed to:
 - allow for a flexible inflation rate determined by market demand targeting a particular bonded-stake ratio
 - effect a balance between market liquidity and staked supply

In order to best determine the appropriate market rate for inflation rewards, a
moving change rate is used.  The moving change rate mechanism ensures that if
the % bonded is either over or under the goal %-bonded, the inflation rate will
adjust to further incentivize or disincentivize being bonded, respectively. Setting the goal
%-bonded at less than 100% encourages the network to maintain some non-staked tokens
which should help provide some liquidity.

It can be broken down in the following way: 
 - If the inflation rate is below the goal %-bonded the inflation rate will
   increase until a maximum value is reached
 - If the goal % bonded (67% in Cosmos-Hub) is maintained, then the inflation
   rate will stay constant 
 - If the inflation rate is above the goal %-bonded the inflation rate will
   decrease until a minimum value is reached
