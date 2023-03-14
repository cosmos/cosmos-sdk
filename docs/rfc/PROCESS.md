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

RFC creation is an **iterative** process. Instead of trying to solve all decisions in a single RFC pull request, we MUST firstly understand the problem and collect feedback through a GitHub Issue.

1. Proposals could start with a new GitHub Issue,  be a result of existing Issues or a discussion.

2. An RFC doesn't have to arrive to `main` with an _accepted_ status in a single PR. If the motivation is clear and the solution is sound, we SHOULD be able to merge it and keep a _proposed_ status. It's preferable to have an iterative approach rather than long, not merged Pull Requests.

3. If a _proposed_ ADR is merged, then it should clearly document outstanding issues either in the RFC document notes or in a GitHub Issue.

4. The PR SHOULD always be merged. In the case of a faulty RFC, we still prefer to  merge it with a _rejected_ status. The only time the RFC SHOULD NOT be merged is if the author abandons it.

5. Merged RFCs SHOULD NOT be pruned.

6. If there is consensus and enough feedback, an ADR can be written on 

### RFC status

Status has two components:

```text
{CONSENSUS STATUS}
```
