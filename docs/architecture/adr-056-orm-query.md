# ADR 056: ORM Query Layer

## Changelog

* {date}: {changelog}

## Status

{DRAFT | PROPOSED} Not Implemented

> Please have a look at the [PROCESS](./PROCESS.md#adr-status) page.
> Use DRAFT if the ADR is in a draft stage (draft PR) or PROPOSED if it's in review.

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.

## Context

> This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts. It should clearly explain the problem and motivation that the proposal aims to resolve.
> {context body}

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

### Logical Query Expressions

Internally queries will be processed using logical query expressions. This will allow queries from the different APIs
to target any backends that support these logical query expressions. Both a simple generic gRPC API and a type-safe
GraphQL API are proposed as well as ORM state store and SQL backends.

A query expression can specify:
* the message name of the table being queried
* a where expression, explained below,
* an order by expression which includes 0 or more fields in either ascending or descending order
* a pagination request (`cosmos.base.query.v1beta1.PageQuest`)

#### Where Expressions

A where expression is composed of one or more field expressions. Each field expression may specify one or two
comparison operators on that field, either:
* `eq`, or
* `lt` or `lte` &/or `gt` or `gte`

Other comparison operators may be added in the future such as `neq`, `in`, `nin`, `regex`, etc.
More generic `and` and `or` expressions may also be added in the future, but for
now the goal is to keep the query layer as simple as possible so that the most value
can be provided with the least effot. Between the two goals of improving the query language and
adding support for more backends and improving performance, the latter should
be preferred.

### Built-in Query Planner

The built-in query planner which targets ORM state directly should be as simple as possible without being 100%
naive. The proposed approach is to use a simple rule-based query planner
that chooses a single index to perform queries on based on a few heuristics
without comparing alternate paths:

* if an order by clause is specified, then the longest index which is a prefix of the
order by clause is used, or the primary key as default
* if an order by clause is not specified, then the index which matches the most
number of provided field expressions is used, or the primary key as default
* given the index that is being used a range query is built using all of the field
expressions specified in the index up to the first unspecified field
* all of the remaining fields expressions are turned into a filter expression 
* any remaining fields from the order by clause are turned into a sort operator
on the result of the range query on the index
* pagination is applied at the end

### SQL Indexer and Query Translator

In order to provide more optimized queries with the ability to add ad hoc indexed based on client needs, an SQL indexer
and query translator will be provided.

The indexer will listen to ORM updates and index them into any SQL database using the [GORM](https://gorm.io) library.

The query translator will take logical query expressions and turn them into SQL `SELECT` expressions. This will
allow queries to be targeted against either the state or an SQL database whichever is available.

### Generic gRPC API

A generic gRPC API which does not attempt to provide any sophisticated type safety is proposed. The SDK currently
provides a good gRPC story and the intention is to continue this support. A more type safe query layer in gRPC could
be provided by code generating .proto files and this could be added in the future, but a type-safe GraphQL approach
is seem as likely more beneficial for sophisticated app development.

### GraphQL API

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
