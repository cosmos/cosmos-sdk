# Upgrading Cosmos SDK

 This guide provides instructions for upgrading to specific versions of Cosmos SDK.

 ## Unreleased 

### Modules

- The `x/param` module has been depreacted in favour of each module housing and providing way to modify their parameters. Each module that has parameters that are changable during runtime have an authority, the authority can be a module or user account. The Cosmos-SDK team recommends migrating modules away from using the param module. An example of how this could look like can be found [here](https://github.com/cosmos/cosmos-sdk/pull/12363). 
 		-	The Param module will be maintained until February 18, 2022. At this point the module will reach end of life and be removed from the cosmos-sdk.
