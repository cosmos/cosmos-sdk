## Rest API

**File: [`x/simple_governance/client/rest/simple_governance.goo`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/client/rest/simple_governance.go)**

The Rest Server, also called [Light-Client Daemon (LCD)](https://github.com/cosmos/cosmos-sdk/tree/master/client/lcd), provides support for **HTTP queries**.

________________________________________________________

USER INTERFACE <=======> REST SERVER <=======> FULL-NODE

________________________________________________________

It allows end-users that do not want to run full-nodes themselves to interract with the chain. The LCD can be configured to perform **Light-Client verification** via the flag `--trust-node`, which can be set to `true` or `false`.

- If *light-client verification* is enabled, the Rest Server acts as a light-client and needs to be run on the end-user's machine. It allows them to interract with the chain in a trustless way without having to store the whole chain locally.

- If *light-client verification* is disabled, the Rest Server acts as a simple relayer for HTTP calls. In this setting, the Rest server needs not be run on the end-user's machine. Instead, it will probably be run by the same entity that operates the full-node the server connects to. This mode is useful if end-users trust the full-node operator and do not want to store anything locally.

Now, let us define endpoints that will be available for users to query through HTTP requests. These endpoints will be defined in a `simple_governance.go` file stored in the `rest` folder.

| Method | URL                             | Description                                                 |
|--------|---------------------------------|-------------------------------------------------------------|
| GET    | /proposals                      | Range query to get all submitted proposals                  |
| POST   | /proposals                      | Submit a new proposal                                       |
| GET    | /proposals/{id}                 | Returns a proposal given its ID                             |
| GET    | /proposals/{id}/votes           | Range query to get all the votes casted on a given proposal |
| POST   | /proposals/{id}/votes           | Cast a vote on a given proposal                             |
| GET    | /proposals/{id}/votes/{address} | Returns the vote of a given address on a given proposal     |

It is the job of module developers to provide sensible endpoints so that front-end developers and service providers can properly interact with it.

Additionaly, here is a [link](https://hackernoon.com/restful-api-designing-guidelines-the-best-practices-60e1d954e7c9) for REST APIs best practices.