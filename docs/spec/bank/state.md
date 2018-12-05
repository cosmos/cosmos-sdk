## State

Presently, the bank module has no inherent state — it simply reads and writes accounts using the `AccountKeeper` from the `auth` module.

This implementation choice is intended to minimize necessary state reads/writes, since we expect most transactions to involve coin amounts (for fees), so storing coin data in the account saves reading it separately.
