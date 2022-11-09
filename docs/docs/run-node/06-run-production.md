---
sidebar_position: 1
---

# Running in Production

:::note Synopsis
This section describes how to securely run a node in a public setting and/or on a mainnet on one of the many cosmos-sdk public blockchains. 
:::

When operating a node, full node or validator, in production it is important to set your server up securely. 

:::note
There are many different ways to secure a server and your node, the described steps here is one way.
:::

:::note
This walkthrough assumes the underlying operating system is ubuntu. 
:::
## Sever Setup

### User

When creating a server most times it is created as user `root`. This user has heightened privileges on the server. When operating a node, it is recommended to not run your node as the root user.  

1. Create a new user

```bash
sudo adduser change_me
```

2. We want to allow this user to perform sudo tasks

```bash
sudo usermod -aG sudo change_me
```

Now when logging into the server, the non `root` user can be used. 

### Firewall

Nodes should not have all ports open to the public, this is a simple way to get DDOS'd. Secondly it is recommended by [Tendermint](github.com/tendermint/tendermint) to never expose ports that are not required to operate a node. 

When setting up a firewall there are a few ports that can be open when operating a Cosmos SDK node. There is the Tendermint json-RPC, promethues, p2p, remote signer and Cosmos SDK GRPC and REST. If the node is being operated as a node that does not offer endpoints to be used for submission or querying then a max of three endpoints are needed.

Most, if not all servers come equipped with [ufw](https://wiki.ubuntuusers.de/ufw/). Ufw will be used in this tutorial. 

1. Reset UFW to disallow all incoming connections and allow outgoing

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
```

2. Lets make sure that port 22 (ssh) stays open. 

```bash
sudo ufw allow ssh
```

or 

```bash
sudo ufw allow 22
```
Both of the above commands are the same. 

3. Allow Port 26656 (tendermint p2p port). If the node has a modified p2p port then that port must be used here.

```bash
sudo ufw allow 26656/tcp
```

4. Allow port 26660 (tendermint [prometheus](https://prometheus.io)). This acts as the applications monitoring port as well. 

```bash
sudo ufw allow 26660/tcp
```

5. IF the node which is being setup would like to expose Tendermints jsonRPC and Cosmos SDK GRPC and REST then follow this step. (Optional)

##### Tendermint JsonRPC

```bash
sudo ufw allow 26657/tcp
```

##### Cosmos SDK GRPC

```bash
sudo ufw allow 9090/tcp
```

##### Cosmos SDK REST

```bash
sudo ufw allow 1317/tcp
```

6. Lastly enabling ufw

```bash
sudo ufw enable
```

### Signing

If the node that is being started is a validator there are multiple ways a validator could sign blocks. 

#### File

File based signing is the simplest and default approach. This approach works by storing the consensus key, generated on initialization, to sign blocks. This approach is only as safe as your server setup as if the server is compromised so is your key.  This key is located in the `config/priv_val_key.json` directory generated on initialization
