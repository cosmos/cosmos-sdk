### `cosmos/cosmos-sdk` Release Process

- [ ] 1. Decide on release designation (are we doing a patch, or minor version bump) and start a P.R. for the release
- [ ] 2. Add commits/PRs that are desired for this release **that haven’t already been added to develop**
- [ ] 3. Merge items in `PENDING.md` into the `CHANGELOG.md`. While doing this make sure that each entry contains links to issues/PRs for each item
- [ ] 4. Summarize breaking API changes section under “Breaking Changes” section to the `CHANGELOG.md` to bring attention to any breaking API changes that affect RPC consumers.
- [ ] 5. Tag the commit `{{ .Release.Name }}-rcN`
- [ ] 6. Kick off 1 day of automated fuzz testing
- [ ] 7. Release Lead assigns 2 people to perform [buddy testing script](/docs/RELEASE_TEST_SCRIPT.md) and update the relevant documentation
- [ ] 8. If errors are found in either #6 or #7 go back to #2 (*NOTE*: be sure to increment the `rcN`)
- [ ] 9. After #6 and #7 have successfully completed then merge the release PR and push the final release tag
