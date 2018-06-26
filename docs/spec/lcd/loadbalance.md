# Load Balancing Module 

This module is designed for the load balancing problem. Because the CosmosSDK LCD will connect with many servers in the future. It is bound to need a load balancing function that improves the distribution of workloads across many full nodes.

## Design
This module need combine with client to realize the real load balancing. It can embed into the `HTTP`(Loc:"github.com/tendermint/tendermint/rpc/lib/client/httpclient.go"). In other words，we realize the new httpclient based on `HTTP`.

From:

```go
type HTTP struct {
	remote string
	rpc    *rpcclient.JSONRPCClient
	*WSEvents
}
```
To:

```go
type HTTPLoadBalance struct {
	mgr BalanceMgr
	rpcs   map[string]*rpcclient.JSONRPCClient
	currpc *rpcclient.JSONRPCClient
	*WSEvents
}
```
* `mgr`：It is the realization of the load balancing and contains all the addresses. `mgr.GetAddr(name string)` get the remote address to serve the current request by the load balancing algorithm that is decided by the `name`.
* `rpcs`：It mapped the string of the remote address to the `rpcclient.JSONRPCClient`.
* `rpc`: the current working rpcclient

### The Diagram of LCD RPC WorkFlow with LoadBalance
![The Diagram of LCD RPC WorkFlow](pics/loadbalanceDiagram.png)

In the above sequence diagram, application calls the `Request()`, and LCD finally call the `HTTP.Request()` through the SecureClient `Wrapper`. In every `HTTP.Request()`, `Getclient()` selects the current working rpcclient by the load balancing algorithm,then run the `JSONRPCClient.Call()` to request from the Full Node, finally `UpdateClient()` updates the weight of the current rpcclient according to the status that is returned by the full node. The `GetAddr()` and `UpdateAddrWeight()` are realized  in the load balancing module.

There are some abilities to do:

* Add the Remote Address
* Delete the Remote Address
* Update the weights of the addresses


### Load balancing Strategies
We can design some strategies like nginx to combine the different load balancing algorithms to get the final remote. We can also get the status of the remote server to add or delete the addresses and update weights of the addresses.

In a word，it can make the entire LCD work more effective in actual conditions.
We are working this module independently in this [Github Repository](https://github.com/MrXJC/GoLoadBalance).
## Interface And Type

### Balancer
>in `balance/balance.go`

This interface `Balancer`is the core of the package. Every load balancing algorithm should realize it,and it defined two interfaces.

* `init` initialize the balancer, assigns the variables which `DoBalance` needs.
* `DoBalance` load balance the full node addresses according to the current situation. 

```go
package balance

type  Balancer interface {
	 init(NodeAddrs)
     DoBalance(NodeAddrs) (*NodeAddr,int,error)
}
```

### NodeAddr

>in `balance/types.go`

* host: ip address
* port: the number of port
* weight: the weight of this full node address,default:1

This NodeAddr is the base struct of the address.

```go
type NodeAddr struct{
	host string
	port int
	weight int
}
func (p *NodeAddr) GetHost() string 
func (p *NodeAddr) GetPort() int 
func (p *NodeAddr) GetWeight() int 
func (p *NodeAddr) updateWeight(weight int)
```

The `weight` is the important factor that schedules which full node the LCD calls. The weight can be changed by the information from the full node. So we have the function `updateWegiht`.

### NodeAddrs
>in `balance/types.go`

`NodeAddrs` is the list of the full node address. This is the member variable in the BalanceManager(`BalancerMgr`)

```go
type NodeAddrs []*NodeAddr
```

## Load Balancing Algorithm
### Random
>in `balance/random.go`

Random Algorithm selects a remote address randomly to process the request. The probability of them being selected is the same. 

### RandomWeight
>in `balance/random.go`

RandomWeight Algorithm also selects a remote address randomly to process the request. But the higher the weight, the greater the probability.

### RoundRobin
>in `balance/roundrobin.go`

RoundRobin Algorithm selects a remote address orderly. Every remote address have the same probability to be selected.

### RoundRobinWeight
>in `balance/roundrobin.go`

RoundRobinWeight Algorthm selects a remote address orderly. But every remote address have different probability to be selected which are determined by their weight.

### Hash
> **Todo**

## Load Balancing Manager
### BalanceMgr
>in `balance/manager.go`

* addrs: the set of the remote full node addresses
* balancers: map the string of balancer name to the specific balancer
* change: record whether the machine reinitialize after the `addrs` changes

`BalanceMgr` is the manager of many balancer. It is the access of load balancing. Its main function is to maintain the `NodeAddrs` and to call the specific load balancing algorithm above.

```go
type BalanceMgr struct{
	addrs NodeAddrs
	balancers map[string]Balancer
	change map[string]bool
}
func (p *BalanceMgr) RegisterBalancer(name string,balancer Balancer)
func (p *BalanceMgr) updateBalancer(name string)
func (p *BalanceMgr) AddNodeAddr(addr *NodeAddr)
func (p *BalanceMgr) DeleteNodeAddr(i int)
func (p *BalanceMgr) UpdateWeightNodeAddr(i int,weight int)

func (p *BalanceMgr) GetAddr(name string)(*NodeAddr,int,error) {
   # if addrs change,update the balancer which we use.
	if p.change[name]{
		p.updateBalancer(name)
	}
   # get the balancer by name
	balancer := p.balancers[name]
	
	# use the load balancing algorithm
	addr,index,err := balancer.DoBalance(p.addrs)
	return addr,index,err
}
```

* `RegisterBalancer`: register the basic balancer implementing the `Balancer` interface and initialize them.
* `updateBalancer`: update the specific balancer after the `addrs` change.
* `AddNodeAddr`: add the remote address and set all the values of the `change` to true.
* `DeleteNodeAddr`: delete the remote address and set all the values of the `change` to true.
* `UpdateWeightNodeAddr`: update the weight of the remote address and set all the values of the `change` to true.
* `GetAddr`:select the address by the balancer the `name` decides.  



