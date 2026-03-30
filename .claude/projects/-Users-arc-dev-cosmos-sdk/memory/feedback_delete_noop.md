---
name: Delete no-op code
description: User prefers deleting dead/no-op code rather than documenting it
type: feedback
---

Delete no-op/placeholder code rather than adding doc comments to it. If a function is a no-op stub, remove the file and function entirely instead of documenting it.

**Why:** Documenting dead code gives it false legitimacy; it's better to just remove it.

**How to apply:** When reviewing code and finding stubs, empty implementations, or unused variables/functions, delete them rather than adding documentation.