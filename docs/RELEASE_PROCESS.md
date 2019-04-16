### `cosmos/cosmos-sdk` Release Process

Unless otherwise stated, all of the following tasks will be undertaken by the release manager/coordinator:

- [ ] 1. Decide on release designation (are we doing a patch, or minor version bump)
- [ ] 2. Ensure that all commits/PRs which are destined for this release are merged to the `master` branch
- [ ] 3. Create the release candidate branch (going forward known as **RC**) and ensure it's protected against pushing from anyone except the release manager/coordinator. **no PRs targeting this branch should be merged unless exceptional circumstances arise**
- [ ] 4. On the `RC` branch, merge items in `PENDING.md` into the `CHANGELOG.md`. While doing this, make sure that each entry contains links to issues/PRs for each item
- [ ] 5. Summarize breaking API changes section under “Breaking Changes” section to the `CHANGELOG.md` to bring attention to any breaking API changes that affect RPC consumers.
- [ ] 6. Kick off a large round of simulation testing (e.g. 400 seeds for 2k blocks) 
- [ ] 7. If errors are found in either #6 or #7 go back to #2 (*NOTE*: be sure to increment the `rcN`)
- [ ] 8. After #6 and #7 have successfully completed create the release branch from the `RC` branch (this will trigger the automated relase process which will build binaries, tag the release and push to github)
- [ ] 9. Merge the release branch to `master` and delete the `RC` branch
