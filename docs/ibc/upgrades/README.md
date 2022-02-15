<!--
order: false
parent:
  order: 3
-->

### Upgrading IBC Chains Overview

This directory contains information on how to upgrade an IBC chain without breaking counterparty clients and connections.

IBC-connnected chains must be able to upgrade without breaking connections to other chains. Otherwise there would be a massive disincentive towards upgrading and disrupting high-value IBC connections, thus preventing chains in the IBC ecosystem from evolving and improving. Many chain upgrades may be irrelevant to IBC, however some upgrades could potentially break counterparty clients if not handled correctly. Thus, any IBC chain that wishes to perform a IBC-client-breaking upgrade must perform an IBC upgrade in order to allow counterparty clients to securely upgrade to the new light client.

1. The [quick-guide](./quick-guide.md) describes how IBC-connected chains can perform client-breaking upgrades and how relayers can securely upgrade counterparty clients using the SDK.
2. The [developer-guide](./developer-guide.md) is a guide for developers intending to develop IBC client implementations with upgrade functionality.
