### `cosmos/cosmos-sdk` Release Process

- [ ] 1. Decide on release designation (are we doing a patch, or minor version bump) and start a branch for the release
- [ ] 2. Add commits/PRs that are desired for this release **that haven’t already been added to develop**
- [ ] 3. Merge items in `PENDING.md` into the `CHANGELOG.md`. While doing this make sure that each entry contains links to issues/PRs for each item
- [ ] 4. Summarize breaking API changes section under “Breaking Changes” section to the `CHANGELOG.md` to bring attention to any breaking API changes that affect RPC consumers.
- [ ] 5. Tag the commit `git tag -a { .Release.Name }-rcN -m 'Release { .Release.Name }'`
- [ ] 6. Open PRs for both `master` and `develop`. From now onwards ***no additional PRs targeting develop should be merged**. Additional commits can be added on top of the release branch though.
- [ ] 7. Ensure both `master` and `develop` PRs ***pass tests***.
- [ ] 8. Kick off 1 day of automated fuzz testing
- [ ] 9. Release Lead assigns 2 people to perform [buddy testing script](/docs/RELEASE_TEST_SCRIPT.md) and update the relevant documentation
- [ ] 10. If errors are found in either #6 or #7 go back to #2 (*NOTE*: be sure to increment the `rcN`)
- [ ] 11. After #6 and #7 have successfully completed then merge the release PR create the final release annotated tag:
 - `git tag -a -m { .Release.Name } 'Release { .Release.Name }'
 - Merge **the release tag** to both `master` and `develop` so that both branches sit on top of the same commit: `branches='master develop' ; for b in $branches ; do git checkout $b ; git merge { .Release.Name } ; git push $b`.
 Alternatively merge both the aforementioned release PRs.
 - Push the final annotated release tag: `git push --tags`
