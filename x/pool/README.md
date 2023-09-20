---
sidebar_position: 1
---

# `x/pool`

Functionality to handle community pool funds. This provides a separate module account for community pool making it easier to track the pool assets. We no longer track community pool assets in distribution module, but instead in this pool module. Funds are migrated from the distribution module's community pool to pool's module account.