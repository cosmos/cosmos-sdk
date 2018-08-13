## Future Improvements

### Power Change

Within the current implementation all power changes ever made are indefinitely stored
within the current state. In the future this state should be trimmed on an epoch basis. Delegators
which will have not withdrawn their fees will be penalized in some way, depending on what is 
computationally feasible this may include:
 - burning non-withdrawn fees
 - requiring more expensive withdrawal costs which include proofs from archive nodes of historical state

In addition or as an alternative it may make sense to implement a "rolling" epoch which cycles through 
all the delegators in small groups (for example 5 delegators per block) and just runs the withdrawal transaction 
at standard rates and takes transaction fees from the withdrawal amount.  


