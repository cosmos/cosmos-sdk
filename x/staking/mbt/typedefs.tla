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

SANY_please_dont_forget_me == TRUE
====
