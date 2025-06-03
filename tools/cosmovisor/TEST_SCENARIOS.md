Ideally, scenarios should run in process with args passed manually to cobra so that we can:
- code coverage
- use the debugger

## Scenarios
- basic: start node, get upgrade-info.json, upgrade
- manual upgrade added while running:
  - start node
  - get upgrade-info.json.batch
  - restart with halt height
  - reach halt height
  - upgrade
- manual upgrade added while running but get upgrade-info.json:
  - start node
  - get upgrade-info.json.batch
  - restart with halt height
  - get upgrade-info.json
  - upgrade
  - restart with halt height
  - reach halt height
  - upgrade
- start with halt height