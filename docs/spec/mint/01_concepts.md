# Concepts

## The Minting Mechanism

The minting mechanism was designed to:
 - allow for a market evaluation of the inflation rate
 - optimize between market liquidity and staked supply

In order to best determine the appropriate market rate for inflation rewards, a
moving change rate is used.  The moving change rate mechanism ensures that if
the % bonded is either over of under the goal %-bonded, the inflation rate will
adjust to further incentivize or incentive being bonded. Setting the goal
%-bonded at less than 100% encourages the network to maintain some non-staked tokens
which presumably encourages liquidity. 

It can be broken down in the following way: 
 - If the inflation rate is below the goal %-bonded the inflation rate will
   increase until a maximum cap is reached
 - If the goal % bonded (67% in Cosmos-Hub) is maintained, then the inflation
   rate will stay constant 
 - If the inflation rate is above the goal %-bonded the inflation rate will
   decrease until a minimum cap is reached

