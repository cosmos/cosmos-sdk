<!--
The default pull request template is for types feat, fix, or refactor.
For other templates, add one of the following parameters to the url:
- template=docs.md
- template=other.md
-->

## Description

Closes: #XXXX

<!-- Add a description of the changes that this PR introduces and the files that
are the most critical to review. -->

---

### Author Checklist

*All items are required. Please add a note to the item if the item is not applicable and
please add links to any relevant follow up issues.*

I have...

* [ ] included the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
* [ ] added `!` to the type prefix if API or client breaking change
* [ ] targeted the correct branch (see [PR Targeting](https://github.com/cosmos/cosmos-sdk/blob/main/CONTRIBUTING.md#pr-targeting))
* [ ] provided a link to the relevant issue or specification
* [ ] followed the guidelines for [building modules](https://github.com/cosmos/cosmos-sdk/blob/main/docs/docs/building-modules)
* [ ] included the necessary unit and integration [tests](https://github.com/cosmos/cosmos-sdk/blob/main/CONTRIBUTING.md#testing)
* [ ] added a changelog entry to `CHANGELOG.md`
* [ ] included comments for [documenting Go code](https://blog.golang.org/godoc)
* [ ] updated the relevant documentation or specification
* [ ] reviewed "Files changed" and left comments if necessary
* [ ] run `make lint` and `make test`
* [ ] confirmed all CI checks have passed

### Reviewers Checklist

*All items are required. Please add a note if the item is not applicable and please add
your handle next to the items reviewed if you only reviewed selected items.*

I have...

* [ ] confirmed the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
* [ ] confirmed `!` in the type prefix if API or client breaking change
* [ ] confirmed all author checklist items have been addressed 
* [ ] reviewed state machine logic
* [ ] reviewed API design and naming
* [ ] reviewed documentation is accurate
* [ ] reviewed tests and test coverage
* [ ] manually tested (if applicable)
