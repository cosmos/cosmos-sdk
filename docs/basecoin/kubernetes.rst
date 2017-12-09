Basecoin Kubernetes
===================

*Note:* This guide is based from - and has been updated - the `origin Medium article <https://blog.cosmos.network/get-your-cryptocurrency-up-and-running-in-10-minutes-9ae4c703eb1c>`__.

So, your Tendermint application is finally ready and you want to
distribute it and run it on several machines, or just run it locally by
creating a dozen Docker containers. We have created
`mintnet-kubernetes <https://github.com/tendermint/tools/tree/master/mintnet-kubernetes>`__
to help you achieve this goal as fast as possible. Note that it should
be primarily used for testing purposes or for tightly-defined chains
operated by a single stakeholder (see `the security
precautions <https://github.com/tendermint/tools/tree/master/mintnet-kubernetes#security>`__).
If you want to launch an application with many stakeholders, consider
using `our set of
Ansible <https://github.com/tendermint/tools/tree/master/ansible-tendermint>`__
scripts to deploy Tendermint.

``mintnet-kubernetes`` is a configuration file for
`Kubernetes <https://kubernetes.io/>`__.

*Kubernetes is an open-source system for automating deployment, scaling,
and management of containerized applications.*

If you had never heard of it, it won’t hurt to read `What is
Kubernetes? <https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/>`__
and go through their `interactive
lessons <https://kubernetes.io/docs/tutorials/kubernetes-basics/>`__. It
won't take long, I promise.

There are several ways to install a Kubernetes cluster:

-  a local Docker-based solution using
   `Minikube <https://github.com/kubernetes/minikube>`__
-  hosted solutions using a Web UI or CLI (e.g. GCE)
-  turn-key cloud solutions (e.g. AWS using `Kubernetes
   Operations <https://github.com/kubernetes/kops/blob/master/docs/aws.md>`__;
   AWS, Azure, GCE or bare metal using
   `Kargo <https://kubernetes.io/docs/getting-started-guides/kargo/>`__)
-  custom solutions (e.g. Linux machines using
   `kubeadm <https://kubernetes.io/docs/getting-started-guides/kubeadm/>`__)

If you just want to play with your application, choose a local
installation with Minikube. If you want to run it in the cloud or on
bare metal, refer to `Picking the Right
Solution <https://kubernetes.io/docs/getting-started-guides>`__, taking
into account the cost, safety, reliability, and configuration options of
those solutions.

Further, we will assume that you have a standard 64-bit Linux desktop
with `VirtualBox <https://www.virtualbox.org/wiki/Downloads>`__ or
`KVM <http://www.linux-kvm.org/>`__ installed.

Install kubectl
---------------

::

    curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x ./kubectl && sudo mv ./kubectl /usr/local/bin/kubectl

For Windows or OS X, please check out `Installing and Setting Up
kubectl <https://kubernetes.io/docs/tasks/kubectl/install/>`__ guide.

Install Minikube
----------------

::

    curl -Lo minikube https://storage.googleapis.com/minikube/releases/v0.19.0/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/

For Windows or OS X, please check out
`Releases <https://github.com/kubernetes/minikube/releases>`__ page.

Start Minikube
--------------

::

    minikube start

Configure your app
------------------

Get the config file:

::

    curl -Lo app.yaml https://raw.githubusercontent.com/tendermint/tools/master/mintnet-kubernetes/examples/basecoin/app.yaml

Open it in the editor of your choice:

::

$EDITOR app.yaml

and configure the parameters required for your application.

The Cosmos SDK will help us create our own
currency, extensible with plugins. Writing plugins is out of scope of
this article, but you can read about it
`here <https://github.com/cosmos/cosmos-sdk/blob/master/docs/guide/basecoin-plugins.md>`__.

Kubernetes DSL (Domain Specific Language) can be difficult for the
beginner to grasp. Fortunately, we will need to change only a small
piece of code.

The most important thing is the application’s ``genesis.json`` file. It
defines the initial distribution of assets. We have 4 nodes by default
(``replicas: 4``\ in StatefulSet). Generally, it is a good idea to split
assets evenly between them and have some large number for everyone.

Let’s change denom to “MyAwesomeCoin”. You are welcome to pick any other
name you like.

::
    
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: app-config
    data:
      genesis.json: |-
        {
          "chain_id": "chain-tTH4mi",
          "app_options": {
            "accounts": [
              {
                "pub_key": "tm-0",
                "coins": [
                  {
                    "denom": "MyAwesomeCoin",
                    "amount": 1000000000
                  }
                ]
              },
              {
                "pub_key": "tm-1",
                "coins": [
                  {
                    "denom": "MyAwesomeCoin",
                    "amount": 1000000000
                  }
                ]
              },
              {
                "pub_key": "tm-2",
                "coins": [
                  {
                    "denom": "MyAwesomeCoin",
                    "amount": 1000000000
                  }
                ]
              },
              {
                "pub_key": "tm-3",
                "coins": [
                  {
                    "denom": "MyAwesomeCoin",
                    "amount": 1000000000
                  }
                ]
              }
            ]
          }
        }

Launch your app
---------------

::

    kubectl create -f ./app.yaml

Wait until all of the nodes are running:

::

    kubectl get pods -w -o wide -L tm
    NAME      READY     STATUS    RESTARTS   AGE       IP           NODE       TM
    tm-0      3/3       Running   0          3m        172.17.0.2   minikube   <none>
    tm-1      3/3       Running   0          3m        172.17.0.3   minikube   <none>
    tm-2      3/3       Running   1          3m        172.17.0.4   minikube   <none>
    tm-3      3/3       Running   0          3m        172.17.0.7   minikube   <none>

Let’s check the first account:

::

    ADDR=$(kubectl exec -c app tm-0 -- cat /app/key.json | jq ".address" | tr -d "\"")

    kubectl exec -c app tm-0 -- basecoin account $ADDR
    {"pub_key":{"type":"ed25519","data":"793B7E33EF94132E16534CC9BA59F74944065FA917A98DB68ABA806D219A4529"},"sequence":1,"coins":[{"denom":"MyAwesomeCoin","amount":999999995}]}

Great! Let’s try to send a transaction from the first to the second
account:

::

    RECIPIENT=$(kubectl exec -c app tm-1 -- cat /app/key.json | jq ".address" | tr -d "\"")

    kubectl exec -c app tm-0 -- basecoin tx send --to 0x$RECIPIENT --amount 5MyAwesomeCoin --from /app/key.json --chain_id chain-tTH4mi
    Signed SendTx:
    0100000000000000000104636F696E000000000000000001010114A677E98456071E3240EF0A2E0B80FFE7D36515BF010101066D79636F696E0000000000000005010201E6A038849655CD3C94D06BAC1CA74443D312855A9BC3575311842DF74AF7DB772673DF60F3AE08CC5260AE93DCE4DB588EF24D08768D0DE2752F001DDC1DEE0F0001010114E2AFEA4A193E85A2DBB8668D4EA0DC0B1A6AD63A010101066D79636F696E0000000000000005
    Response: 3D54EECAAE072477E6119C6DF1762168F276F0C1 ;

Checking the first account’s balance we should see 5 coins making their
way into the second account:

::

    kubectl exec -c app tm-0 -- basecoin account $ADDR
    {"pub_key":{"type":"ed25519","data":"793B7E33EF94132E16534CC9BA59F74944065FA917A98DB68ABA806D219A4529"},"sequence":2,"coins":[{"denom":"MyAwesomeCoin","amount":999999990}]}

As you can see, it was fairly simple to launch a new cryptocurrency in a
Kubernetes cluster. Moreover, with Kubernetes you can add new nodes
(pods) with a single command. And using
`federation <https://kubernetes.io/docs/concepts/cluster-administration/federation/>`__,
you can be sure that your currency will stay alive even after loss of
the entire cluster!

Clean up
--------

::

    kubectl delete -f ./app.yaml
    kubectl delete pvc -l app=tm
