---- MODULE staking ----

\* EXTENDS  Integers, FiniteSets, Sequences, TLC, Apalache, typedefs
EXTENDS  Integers, Naturals, FiniteSets, Sequences, TLC, tlcFolds

CONSTANTS 
    \* @type: ADDR;
    v0, 
    \* @type: ADDR;
    v1

NullStr == "NullStr"
NullInt == 0

PARAM_UnbondingTime == 2 \* 	string (time ns) 	"259200000000000"
PARAM_MaxValidators == 1 \*  	uint16 	100
PARAM_KeyMaxEntries == NullInt \*  	uint16 	7 ---- UNUSED
PARAM_HistoricalEntries == NullInt \* 	uint16 	3 ---- UNUSED
PARAM_BondDenom == NullStr \* 	string 	"stake" ---- UNUSED
PARAM_MinCommissionRate == NullStr \*  	string 	"0.000000000000000000" ---- UNUSED

Delegators == {"d0"}
Validators == {v0, v1}
ValidatorPermutationsSymmetry == Permutations(Validators)
Addresses == Delegators \cup Validators
Amounts == 1..3
InitialBalances == 1..4
InitialValidatorBalance == 1
TimeDeltas == 1..2
HeightDeltas == {1}

BONDED == "bonded"
UNBONDING == "unbonding"
UNBONDED == "unbonded"
SUCCEED == "succeed"


Action(nature, timeDelta, heightDelta, delegator, validator, validatorSrc, validatorDst, amount) ==
    [
        nature |-> nature,
        timeDelta |-> timeDelta,
        heightDelta |-> heightDelta,
        delegator |-> delegator,
        validator |-> validator,
        validatorSrc |-> validatorSrc,
        validatorDst |-> validatorDst,
        amount |-> amount
    ]

BeginRedelegateAction(delegator, src, dst, amount) == Action("beginRedelegate", NullInt, NullInt, delegator, NullStr, src, dst, amount)
UndelegateAction(delegator, validator, amount) == Action("undelegate", NullInt, NullInt, delegator, validator, NullStr, NullStr, amount)
DelegateAction(delegator, validator, amount) == Action("delegate", NullInt, NullInt, delegator, validator, NullStr, NullStr, amount)
EndBlockAction(timeDelta, heightDelta) == Action("endBlock", timeDelta, heightDelta, NullStr, NullStr, NullStr, NullStr, NullInt)
InitAction == Action("init", NullInt, NullInt, NullStr, NullStr, NullStr, NullStr, NullInt)


VARIABLES
(*Meta*)
    \* @type: Int;
    steps,
    \* @type: ACTION;
    action,
    \* @type: Str;
    outcome,
(*Model*)
    \* @type: Int;
    blockTime,
    \* @type: Int;
    blockHeight,
    \* @type: ADDR -> Bool; Validator is jailed?
    jailed, 
    \* @type: ADDR -> STATUS; Validator status
    status,
    \* @type: ADDR -> Int;
    tokens, 
    \* @type: ADDR -> Int; If validator unbonding defines height that begun unbonding
    unbondingHeight,
    \* @type: ADDR -> Int; If validator unbonding defines min time to complete unbonding
    unbondingTime,
    \* @type: <<ADDR, ADDR>> -> Int; Shares for a <<delegator, validator>> relationship
    delegation,
    \* @type: Set(ADDR);
    validatorQ,
    \* @type: Set(UNDELEGATION);
    undelegationQ,
    \* @type: Set(REDELEGATION);
    redelegationQ,
    \* @type: Set(ADDR);
    lastValSet 

SumShares(validator) ==
    LET
    Sum(acc, key) == acc + delegation[key]
    IN
    \* Include mandatory self-delegation
    FoldSet(Sum, 0, (Delegators \cup {validator}) \X {validator})

IsBonded(validator) == status[validator] = BONDED
IsNotBonded(validator) == status[validator] # BONDED
IsUnbonding(validator) == status[validator] = UNBONDING
IsUnbonded(validator) == status[validator] = UNBONDED

InvalidExRate(validator) == tokens[validator] = 0 /\ 0 < SumShares(validator)

HasReceivingRedelegation(delegator, src) ==
    \E red \in redelegationQ :
    /\ red.delegator = delegator
    /\ red.validatorDst = src

\* @type: (Set(UNDELEGATION), UNDELEGATION) => Set(UNDELEGATION);
MergedUndelegationQ(q, und) == 
    LET 
    \* @type: (UNDELEGATION) => Bool;
    Match(other) == 
            /\ und.delegator = other.delegator
            /\ und.validator = other.validator
            /\ und.creationHeight = other.creationHeight
            /\ und.completionTime = other.completionTime
    IN
    LET Merged ==
        LET 
        \* @type: (UNDELEGATION, UNDELEGATION) => UNDELEGATION;
        Combine(acc, other) ==
            IF Match(other) THEN [acc EXCEPT 
                !.balance = @ + other.balance
            ] ELSE acc
        IN FoldSet(Combine, und, q)
    IN
    {e \in q : ~Match(e)} \cup {Merged}

\* @type: (Set(REDELEGATION), REDELEGATION) => Set(REDELEGATION);
MergedRedelegationQ(q, red) == 
    LET 
    \* @type: (REDELEGATION) => Bool;
    Match(other) == 
            /\ red.delegator = other.delegator
            /\ red.validatorSrc = other.validatorSrc
            /\ red.validatorDst = other.validatorDst
            /\ red.creationHeight = other.creationHeight
            /\ red.completionTime = other.completionTime
    IN
    LET Merged ==
        LET 
        \* @type: (REDELEGATION, REDELEGATION) => REDELEGATION;
        Combine(acc, other) ==
            IF Match(other) THEN [acc EXCEPT 
                !.initialBalance = @ + other.initialBalance,
                !.sharesDst = @ + other.sharesDst
            ] ELSE acc
        IN FoldSet(Combine, red, q)
    IN
    {e \in q : ~Match(e)} \cup {Merged}

Init == 
    /\ steps = 0
    /\ action = InitAction
    /\ outcome = NullStr
    /\ blockTime = 0
    /\ blockHeight = 0
    /\ jailed = [v \in Validators |-> FALSE] 
    /\ status = [v \in Validators |-> UNBONDED]
    /\ tokens \in {f \in [Addresses -> InitialBalances] : \A v \in Validators : f[v] = InitialValidatorBalance}
    /\ unbondingHeight = [v \in Validators |-> NullInt]
    /\ unbondingTime = [v \in Validators |-> NullInt]
    /\ delegation = [pair \in (Delegators \X Validators) |-> 0]
                    @@ [ pair \in {p \in Validators \X Validators : p[1] = p[2] } |-> InitialValidatorBalance]
    /\ validatorQ = {}
    /\ undelegationQ = {}
    /\ redelegationQ = {}
    /\ lastValSet = {}

CandidateValidatorSets ==
    LET
        Combine(best_so_far, candidate) ==
        LET
            Power(ss) == 
            LET
                Sum(acc, v) == acc + tokens[v]
            IN
            FoldSet(Sum, 0, ss)
        IN 
        CASE 
            Power(best_so_far) <= Power(candidate) /\ Cardinality(candidate) <= PARAM_MaxValidators -> candidate
          [] OTHER -> best_so_far

    \* V0 has lexicographical priority over V1
    IN {FoldSeq(Combine, {}, <<{v1}, {v0}, {v0, v1}>>)}

EndBlock(t, h) ==
    LET
    IsValid(s) == \A v \in s: (~jailed[v] /\ 0 < tokens[v])
    ExpiredValidators == {v \in validatorQ : unbondingTime[v] <= blockTime /\ unbondingHeight[v] <= blockHeight}
    ExpiredUndelegations == {e \in undelegationQ : e.completionTime <= blockTime}
    ExpiredRedelegations == {e \in redelegationQ : e.completionTime <= blockTime}
    IN
    \E newValSet \in CandidateValidatorSets:
    /\ IsValid(newValSet)
    /\ outcome' = SUCCEED
    /\ blockTime' = blockTime + t 
    /\ blockHeight' = blockHeight + h
    /\ UNCHANGED jailed
    /\ status' = [v \in Validators |->
            CASE v \in newValSet -> BONDED
              [] v \in lastValSet \ newValSet -> UNBONDING
              [] v \in ExpiredValidators \ newValSet -> UNBONDED
              [] OTHER -> status[v]
        ]
    /\ tokens' = [
            a \in Addresses |->
                LET
                    Refund ==
                    LET
                    \* @type: (Int, UNDELEGATION) => Int;
                    Sum(acc, und) == acc + und.balance
                    IN FoldSet(Sum, 0, {und \in ExpiredUndelegations : a = und.delegator})
                IN tokens[a] + Refund
        ]
    /\ unbondingHeight' = [ v \in Validators |-> 
            CASE v \in lastValSet \ newValSet -> blockHeight \* Added to validatorQ
              [] v \in (ExpiredValidators \cup newValSet) -> NullInt \* Removed from validatorQ
              [] OTHER -> unbondingHeight[v]
        ]
    /\ unbondingTime' = [ v \in Validators |-> 
            CASE v \in lastValSet \ newValSet -> blockTime + PARAM_UnbondingTime \* Added to validatorQ
              [] v \in (ExpiredValidators \cup newValSet) -> NullInt \* Removed from validatorQ
              [] OTHER -> unbondingTime[v]
        ]
    /\ UNCHANGED delegation
    /\ validatorQ' = (validatorQ \cup lastValSet) \ (ExpiredValidators \cup newValSet)
    /\ undelegationQ' = undelegationQ \ ExpiredUndelegations 
    /\ redelegationQ' = redelegationQ \ ExpiredRedelegations
    /\ lastValSet' = newValSet 


Delegate(delegator, validator, amount) == 
    LET
    Fail(reason) ==
        /\ outcome' = reason
        /\ UNCHANGED blockTime
        /\ UNCHANGED blockHeight
        /\ UNCHANGED jailed 
        /\ UNCHANGED status
        /\ UNCHANGED tokens 
        /\ UNCHANGED unbondingHeight
        /\ UNCHANGED unbondingTime
        /\ UNCHANGED delegation
        /\ UNCHANGED validatorQ
        /\ UNCHANGED undelegationQ
        /\ UNCHANGED redelegationQ
        /\ UNCHANGED lastValSet  
    Succeed ==
        LET
            IssuedShares ==
            IF SumShares(validator) = 0
            THEN amount 
            ELSE (SumShares(validator) * amount ) \div tokens[validator]
        IN
        /\ outcome' = SUCCEED
        /\ UNCHANGED blockTime
        /\ UNCHANGED blockHeight
        /\ UNCHANGED jailed 
        /\ UNCHANGED status
        /\ tokens' = [
                a \in Addresses |->
                CASE a = delegator -> tokens[a] - amount
                  [] a = validator -> tokens[a] + amount
                  [] OTHER -> tokens[a]
            ]
        /\ UNCHANGED unbondingHeight
        /\ UNCHANGED unbondingTime
        /\ delegation' = [delegation EXCEPT ![delegator, validator] = @ + IssuedShares]
        /\ UNCHANGED validatorQ
        /\ UNCHANGED undelegationQ
        /\ UNCHANGED redelegationQ
        /\ UNCHANGED lastValSet
    IN
    CASE InvalidExRate(validator) -> Fail("fail:invalid_ex_rate")
      [] tokens[delegator] < amount -> Fail("fail:insufficient_tokens")
      [] OTHER -> Succeed


    
Undelegate(delegator, validator, amount) == 
    LET
    Shares == 
        IF tokens[validator] = 0 THEN NullInt ELSE
        (SumShares(validator) * amount) \div tokens[validator]
    IN 
    LET
    Fail(reason) ==
        /\ outcome' = reason
        /\ UNCHANGED blockTime
        /\ UNCHANGED blockHeight
        /\ UNCHANGED jailed 
        /\ UNCHANGED status
        /\ UNCHANGED tokens 
        /\ UNCHANGED unbondingHeight
        /\ UNCHANGED unbondingTime
        /\ UNCHANGED delegation
        /\ UNCHANGED validatorQ
        /\ UNCHANGED undelegationQ
        /\ UNCHANGED redelegationQ
        /\ UNCHANGED lastValSet  
    Succeed ==
        LET
            IssuedTokens ==
            IF SumShares(validator) - Shares = 0
            THEN tokens[validator]
            ELSE (Shares * tokens[validator]) \div SumShares(validator)
        IN
        /\ outcome' = SUCCEED
        /\ UNCHANGED blockTime
        /\ UNCHANGED blockHeight
        /\ UNCHANGED jailed 
        /\ UNCHANGED status
        /\ tokens' = [
            a \in Addresses |->
            CASE a = validator -> tokens[a] - IssuedTokens
              [] OTHER -> tokens[a]
            ]
        /\ UNCHANGED unbondingHeight
        /\ UNCHANGED unbondingTime
        /\ delegation' = [delegation EXCEPT ![delegator, validator] = @ - Shares]
        /\ UNCHANGED validatorQ
        /\ undelegationQ' = MergedUndelegationQ(undelegationQ,
                [
                    delegator |-> delegator,
                    validator |-> validator,
                    creationHeight |-> blockHeight,
                    completionTime |-> blockTime + PARAM_UnbondingTime,
                    balance |-> IssuedTokens
                ]
            )
        /\ UNCHANGED redelegationQ
        /\ UNCHANGED lastValSet 
    IN
    CASE delegation[delegator, validator] < Shares -> Fail("fail:insufficient_shares")
      [] tokens[validator] < 1 -> Fail("fail:insufficient_tokens")
      [] OTHER -> Succeed

BeginRedelegate(delegator, src, dst, amount) == 
    LET
    Shares == 
        IF tokens[src] = 0 THEN NullInt ELSE
        (SumShares(src) * amount) \div tokens[src]
    IN 
    LET
    Fail(reason) ==
        /\ outcome' = reason
        /\ UNCHANGED blockTime
        /\ UNCHANGED blockHeight
        /\ UNCHANGED jailed 
        /\ UNCHANGED status
        /\ UNCHANGED tokens 
        /\ UNCHANGED unbondingHeight
        /\ UNCHANGED unbondingTime
        /\ UNCHANGED delegation
        /\ UNCHANGED validatorQ
        /\ UNCHANGED undelegationQ
        /\ UNCHANGED redelegationQ
        /\ UNCHANGED lastValSet  
    Succeed ==
        LET
            IssuedTokens ==
            IF SumShares(src) - Shares = 0
            THEN tokens[src]
            ELSE (Shares * tokens[src]) \div SumShares(src)
        IN
        LET
            IssuedShares ==
            IF SumShares(dst) = 0
            THEN IssuedTokens 
            ELSE (SumShares(dst) * IssuedTokens ) \div tokens[dst]
        IN
        /\ outcome' = SUCCEED
        /\ UNCHANGED blockTime
        /\ UNCHANGED blockHeight
        /\ UNCHANGED jailed 
        /\ UNCHANGED status
        /\ tokens' = [
            a \in Addresses |->
            CASE a = src -> tokens[a] - IssuedTokens
              [] a = dst -> tokens[a] + IssuedTokens
              [] OTHER -> tokens[a]
            ]
        /\ UNCHANGED unbondingHeight
        /\ UNCHANGED unbondingTime
        /\ delegation' = [delegation EXCEPT
                ![delegator, src] = @ - Shares,
                ![delegator, dst] = @ + IssuedShares
            ]
        /\ UNCHANGED validatorQ
        /\ UNCHANGED undelegationQ
        /\ redelegationQ' = 
            CASE IsBonded(src) -> MergedRedelegationQ(redelegationQ,[
                        delegator |-> delegator,
                        validatorSrc |-> src,
                        validatorDst |-> dst,
                        creationHeight |-> blockHeight,
                        completionTime |-> blockTime + PARAM_UnbondingTime,
                        initialBalance |-> IssuedTokens,
                        sharesDst |-> IssuedShares 
                    ]
                )
              [] IsUnbonding(src) -> MergedRedelegationQ(redelegationQ,[
                        delegator |-> delegator,
                        validatorSrc |-> src,
                        validatorDst |-> dst,
                        creationHeight |-> unbondingHeight[src],
                        completionTime |-> unbondingTime[src],
                        initialBalance |-> IssuedTokens,
                        sharesDst |-> IssuedShares 
                    ]
                )
              [] OTHER -> redelegationQ
        /\ UNCHANGED lastValSet 
    IN
    CASE delegation[delegator, src] < Shares -> Fail("fail:insufficient_shares")
      [] tokens[src] = 0 -> Fail("fail:insufficient_tokens") 
      [] InvalidExRate(dst) -> Fail("fail:invalid_ex_rate")
      [] HasReceivingRedelegation(delegator, src) -> Fail("fail:has_receiving_redelegation") 
      [] src = dst -> Fail("fail:redelegate_to_same_validator")
      [] OTHER -> Succeed

Next ==
    /\ steps' = steps + 1
    /\ \/ \E t \in TimeDeltas, h \in HeightDeltas:
          /\ action' = EndBlockAction(t, h)
          /\ EndBlock(t, h)
       \/ \E delegator \in Delegators, validator \in Validators, amount \in Amounts:
          /\ action' = DelegateAction(delegator, validator, amount)
          /\ Delegate(delegator, validator, amount)
       \/ \E delegator \in Delegators, validator \in Validators, amount \in Amounts:
          /\ action' = UndelegateAction(delegator, validator, amount)
          /\ Undelegate(delegator, validator, amount)
       \/ \E delegator \in Delegators, src, dst \in Validators, amount \in Amounts:
          /\ action' = BeginRedelegateAction(delegator, src, dst, amount)
          /\ BeginRedelegate(delegator, src, dst, amount)
    

Sanity0 == action.nature = "endBlock" 
Sanity1 == action.nature = "delegate" 
Sanity2 == action.nature = "undelegate" 
Sanity3 == action.nature = "beginRedelegate" 

STEPS == 7

P0 ==  ~(STEPS < steps /\ action.nature = "beginRedelegate" /\ outcome = "fail:insufficient_shares")
P1 ==  ~(STEPS < steps /\ action.nature = "beginRedelegate" /\ outcome = "fail:has_receiving_redelegation")
P2 ==  ~(STEPS < steps /\ action.nature = "beginRedelegate" /\ outcome = "fail:redelegate_to_same_validator")
P3 ==  ~(STEPS < steps /\ action.nature = "beginRedelegate" /\ outcome = "succeed")
P4 ==  ~(STEPS < steps /\ action.nature = "undelegate" /\ outcome = "fail:insufficient_shares")
P5 ==  ~(STEPS < steps /\ action.nature = "undelegate" /\ outcome = "succeed")
P6 ==  ~(STEPS < steps /\ action.nature = "delegate" /\ outcome = "fail:insufficient_tokens")
P7 ==  ~(STEPS < steps /\ action.nature = "delegate" /\ outcome = "succeed")
P8 == ~(STEPS < steps /\ action.nature = "endBlock" /\ outcome = "succeed")

(*
~~~~~~~~~~~
The following are only relevant if slashing is modeled.
~~~~~~~~~~~
*)
P9 == ~(STEPS < steps /\ action.nature = "beginRedelegate" /\ outcome = "fail:insufficient_tokens")
P10 == ~(STEPS < steps /\ action.nature = "beginRedelegate" /\ outcome = "fail:invalid_ex_rate")
P11 == ~(STEPS < steps /\ action.nature = "undelegate" /\ outcome = "fail:insufficient_tokens")
P12 == ~(STEPS < steps /\ action.nature = "delegate" /\ outcome = "fail:invalid_ex_rate")

View == <<
    \* steps,
    action,
    outcome,
    blockTime,
    blockHeight,
    jailed, 
    status,
    tokens, 
    unbondingHeight,
    unbondingTime,
    delegation,
    validatorQ,
    undelegationQ,
    redelegationQ,
    lastValSet 
    >>

====