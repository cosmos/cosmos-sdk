# ADR Creation Process

1. Copy `adr-template.md` and in a new file with the following name pattern: `adr-next_number-title.md`
1. Create a Draft Pull Request if you want to get an early feedback.
1. Make sure the context and a solution is clear and well documented.
1. Add an entry to the list below in the [README](./README.md) file.
1. Create a Pull Request to propose a new ADR.


## ADR life cycle

ADR creation is an **iterative** process. Instead of trying to solve all decisions in a single ADR pull request, we MUST firstly understand the problem and collect a feedback through a GitHub Issue.

1. Every proposal SHOULD starts with a new GitHub Issue or be a result of existing Issues. The issue should contain just a brief proposal summary.

1. Once the motivation is validated, a GitHub Pull Request (PR) is created with a new document based on the `adr-template.md`.

1. An ADR doesn't have to arrive to `master` with an _accepted_ status in a single PR. If the motivation is clear and the solution is sound, we SHOULD be able to merge it and keep a _proposed_ status. It's preferable to have an iterative approach rather than long, not merged Pull Requests.

1. If a _proposed_ ADR is merged, then it should clearly document outstanding issues either in ADR document notes or in a GitHub Issue.

1. The PR SHOULD always be merged. In the case of a faulty ADR, we still want to  merge it with a _rejected_ status. The only time the ADR SHOULD NOT be merged is if the author abandons it.


### ADR status

```
DRAFT -> PROPOSED -> LAST CALL yyyy-mm-dd -> FINAL | REJECTED -> SUPERSEEDED by ADR-xxx
                  \        |
                   \       |
                    v      v
                     ABANDONED
```


+ `DRAFT`: [optional] an ADR which is work in progress, not being ready for a general review. This is to present an early work and get an early feedback in a Draft Pull Request form.
+ `PROPOSED`: and ADR covering a full solution architecture.
+ `LAST CALL <date for the last call>`: [optional] clear notify that we are close to accept updates
+ `ACCEPTED`: ADR which will represent a currently implemented or to be implemented architecture design.
+ `SUPERSEEDED by ADR-xxx`: ADR which has been superseded by a new ADR.
+ `ABANDONED`: the ADR is no longer pursued by the original authors.


## Language used in ADR

+ The context/background should be written in the present tense.
+ Avoid using a first, personal form.
