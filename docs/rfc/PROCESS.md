# RFC Creation Process

1. Copy the `rfc-template.md` file. Use the following filename pattern: `rfc-next_number-title.md`
2. Create a draft Pull Request if you want to get an early feedback.
3. Make sure the context and a solution is clear and well documented.
4. Add an entry to a list in the [README](./README.md) file.
5. Create a Pull Request to propose a new ADR.

## What is an RFC?

An RFC is a sort of async whiteboarding session. It is meant to replace the need for a distributed team to come together to make a decision. Currently, the Cosmos SDK team and contributors are distributed around the world. The team conducts working groups to have a synchronous discussion and an RFC can be used to capture the discussion for a wider audience to better understand the changes that are coming to the software. 

The main difference the Cosmos SDK is defining as a differentiation between RFC and ADRs is that one is to come to consensus and circulate information about a potential change or feature. An ADR is used if there is already consensus on a feature or change and there is not a need to articulate the change coming to the software. An ADR will articulate the changes and have a lower amount of communication .   

## RFC life cycle

RFC creation is an **iterative** process. An RFC is meant as a distributed colloboration session, it may have many comments and is usually the bi-product of no working group or synchornous communication 

1. Proposals could start with a new GitHub Issue,  be a result of existing Issues or a discussion.

2. An RFC doesn't have to arrive to `main` with an _accepted_ status in a single PR. If the motivation is clear and the solution is sound, we SHOULD be able to merge it and keep a _proposed_ status. It's preferable to have an iterative approach rather than long, not merged Pull Requests.

3. If a _proposed_ RFC is merged, then it should clearly document outstanding issues either in the RFC document notes or in a GitHub Issue.

4. The PR SHOULD always be merged. In the case of a faulty RFC, we still prefer to  merge it with a _rejected_ status. The only time the RFC SHOULD NOT be merged is if the author abandons it.

5. Merged RFCs SHOULD NOT be pruned.

6. If there is consensus and enough feedback then the RFC can be accepted. 

> Note: An RFC is written when there is no working group or team session on the problem. RFC's are meant as a distributed white boarding session. If there is a working group on the proposal there is no need to have an RFC as there is synchornous whiteboarding going on. 

### RFC status

Status has two components:

```text
{CONSENSUS STATUS}
```

#### Consensus Status

```text
DRAFT -> PROPOSED -> LAST CALL yyyy-mm-dd -> ACCEPTED | REJECTED -> SUPERSEDED by ADR-xxx
                  \        |
                   \       |
                    v      v
                     ABANDONED
```

* `DRAFT`: [optional] an ADR which is work in progress, not being ready for a general review. This is to present an early work and get an early feedback in a Draft Pull Request form.
* `PROPOSED`: an ADR covering a full solution architecture and still in the review - project stakeholders haven't reached an agreed yet.
* `LAST CALL <date for the last call>`: [optional] clear notify that we are close to accept updates. Changing a status to `LAST CALL` means that social consensus (of Cosmos SDK maintainers) has been reached and we still want to give it a time to let the community react or analyze.
* `ACCEPTED`: ADR which will represent a currently implemented or to be implemented architecture design.
* `REJECTED`: ADR can go from PROPOSED or ACCEPTED to rejected if the consensus among project stakeholders will decide so.
* `SUPERSEEDED by ADR-xxx`: ADR which has been superseded by a new ADR.
* `ABANDONED`: the ADR is no longer pursued by the original authors.

## Language used in RFC

* The background/goal should be written in the present tense.
* Avoid using a first, personal form.
