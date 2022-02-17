---- MODULE typedefs ---- 

(*

@typeAlias: ADDR = Str;
@typeAlias: STATUS = Str;

@typeAlias: ACTION = [
    nature : Str,
    timeDelta : Int,
    heightDelta : Int, 
    delegator : ADDR,
    validator : ADDR,
    validatorSrc : ADDR,
    validatorDst : ADDR,
    amount: Int
];

@typeAlias: UNDELEGATION = [
    delegator : ADDR,
    validator : ADDR,
    creationHeight : Int,
    completionTime : Int,
    balance : Int
];

@typeAlias: REDELEGATION = [
    delegator : ADDR,
    validatorSrc : ADDR,
    validatorDst : ADDR,
    creationHeight : Int,
    completionTime : Int,
    initialBalance : Int, 
    sharesDst : Int 
];

*) 

(*
type UnbondingDelegationEntry struct {
	// creation_height is the height which the unbonding took place.
	CreationHeight int64 
	// completion_time is the unix time for unbonding completion.
	CompletionTime time.Time 
	// initial_balance defines the tokens initially scheduled to receive at completion.
	InitialBalance github_com_cosmos_cosmos_sdk_types.Int 
	// balance defines the tokens to receive at completion.
	Balance github_com_cosmos_cosmos_sdk_types.Int 
}

type RedelegationEntry struct {
	// creation_height  defines the height which the redelegation took place.
	CreationHeight int64 
	// completion_time defines the unix time for redelegation completion.
	CompletionTime time.Time 
	// initial_balance defines the initial balance when redelegation started.
	InitialBalance github_com_cosmos_cosmos_sdk_types.Int 
	// shares_dst is the amount of destination-validator shares created by redelegation.
	SharesDst github_com_cosmos_cosmos_sdk_types.Dec 
}
*)

SANY_please_dont_forget_me == TRUE
====
