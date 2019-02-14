# 자체 테스트넷 구축하기

해당 문서는 `gaiad` 노드 네트워크를 구축하는 세가지 방법을 제시합니다. 각 모델은 다른 이용 사례에 특화되어 있습니다.

1. 싱글-노드, 로컬, 수동 테스트넷
2. 멀티-노드, 로컬, 자동 테스트넷
3. 멀티-노드, 리모트, 자동 테스트넷

관련 코드는 [네트워크 디렉토리](https://github.com/cosmos/cosmos-sdk/tree/develop/networks)와 하단의 `local`과 `remote` 서브 디렉토리에서 찾으실 수 있습니다.

> 참고: 현재 `remote` 관련 정보는 최신 릴리즈와 호환성이 맞지 않을 수 있으므로 참고하시기 바랍니다.

## 싱글-노드, 로컬, 수동 테스트넷

이 가이드는 하나의 검증인 노드를 로컬 환경에서 운영하는 방식을 알려드립니다. 이런 환경은 테스트/개발 환경을 구축하는데 이용될 수 있습니다.

### 필수 사항

- [gaia 설치](./installation.md)
- [`jq` 설치](https://stedolan.github.io/jq/download/) (선택 사항)

### 제네시스 파일 만들기, 네트워크 시작하기

```bash
# You can run all of these commands from your home directory
cd $HOME

# Initialize the genesis.json file that will help you to bootstrap the network
gaiad init --chain-id=testing testing

# Create a key to hold your validator account
gaiacli keys add validator

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom, the default is staking
gaiad add-genesis-account $(gaiacli keys show validator -a) 1000stake,1000validatortoken

# Generate the transaction that creates your validator
gaiad gentx --name validator

# Add the generated bonding transaction to the genesis file
gaiad collect-gentxs

# Now its safe to start `gaiad`
gaiad start
```

이 셋업은 모든 `gaiad` 정보를  `~/.gaiad`에 저장힙니다. 생성하신 제네시스 파일을 확인하고 싶으시다면 `~/.gaiad/config/genesis.json`에서 확인이 가능합니다. 위의 세팅으로 `gaiacli`가 이용이 가능하며, 토큰(스테이킹/커스텀)이 있는 계정 또한 함께 생성됩니다.

## 멀티 노드, 로컬, 자동 테스트넷

관련 코드 [networks/local 디렉토리](https://github.com/cosmos/cosmos-sdk/tree/develop/networks/local):

### 필수 사항

- [gaia 설치](./installation.md)
- [docker 설치](https://docs.docker.com/engine/installation/)
- [docker-compose 설치](https://docs.docker.com/compose/install/)

### 빌드

`localnet` 커맨드를 운영하기 위한 `gaiad` 바이너리(리눅스)와 `tendermint/gaiadnode` docker 이미지를 생성합니다. 해당 바이너리는 컨테이너에 마운팅 되며 업데이트를 통해 이미지를 리빌드 하실 수 있습니다.

Build the `gaiad` binary (linux) and the `tendermint/gaiadnode` docker image required for running the `localnet` commands. This binary will be mounted into the container and can be updated rebuilding the image, so you only need to build the image once.

```bash
# Work from the SDK repo
cd $GOPATH/src/github.com/cosmos/cosmos-sdk

# Build the linux binary in ./build
make build-linux

# Build tendermint/gaiadnode image
make build-docker-gaiadnode
```

### 테스트넷 실행하기

4개 노드 테스트넷을 실행하기 위해서는:

```
make localnet-start
```

이 커맨드는 4개 노드로 구성되어있는 네트워크를 gaiadnode 이미지를 기반으로 생성합니다. 각 노드의 포트는 하단 테이블에서 확인하실 수 있습니다:


| 노드 ID | P2P 포트 | RPC 포트 |
| --------|-------|------|
| `gaianode0` | `26656` | `26657` |
| `gaianode1` | `26659` | `26660` |
| `gaianode2` | `26661` | `26662` |
| `gaianode3` | `26663` | `26664` |

바이너리를 업데이트 하기 위해서는 리빌드를 하신 후 노드를 재시작 하시면 됩니다:

```
make build-linux localnet-start
```

### 설정

`make localnet-start`는 `gaiad testnet` 명령을 호출하여 4개 노드로 구성된 테스트넷에 필요한 파일을 `./build`에 저장합니다. 이 명령은 `./build` 디렉토리에 다수의 파일을 내보냅니다.


```bash
$ tree -L 2 build/
build/
├── gaiacli
├── gaiad
├── gentxs
│   ├── node0.json
│   ├── node1.json
│   ├── node2.json
│   └── node3.json
├── node0
│   ├── gaiacli
│   │   ├── key_seed.json
│   │   └── keys
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
├── node1
│   ├── gaiacli
│   │   └── key_seed.json
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
├── node2
│   ├── gaiacli
│   │   └── key_seed.json
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
└── node3
    ├── gaiacli
    │   └── key_seed.json
    └── gaiad
        ├── ${LOG:-gaiad.log}
        ├── config
        └── data
```

각 `./build/nodeN` 디렉토리는 각자 컨테이너 안에 있는 `/gaiad`에 마운팅 됩니다.

### 로깅

로그는 각 `./build/nodeN/gaiad/gaia.log`에 저장됩니다. 로그는 docker를 통해서 바로 확인하실 수도 있습니다:

```
docker logs -f gaiadnode0
```

### 키와 계정

`gaiacli`를 이용해 tx를 생성하거나 상태를 쿼리 하시려면, 특정 노드의 `gaiacli` 디렉토리를 `home`처럼 이용하시면 됩니다. 예를들어: 


```shell
gaiacli keys list --home ./build/node0/gaiacli
```
이제 계정이 존재하니 추가로 새로운 계정을 만들고 계정들에게 토큰을 전송할 수 있습니다.

::: tip
**참고**: 각 노드의 시드는 `./build/nodeN/gaiacli/key_seed.json`에서 확인이 가능하며 `gaiacli keys add --restore` 명령을 통해 CLI로 복원될 수 있습니다.
:::

### 특수 바이너리

다수의 이름을 가진 다수의 바이너리를 소유하신 경우, 어떤 바이너리의 환경 변수(environment variable)를 기준으로 실행할지 선택할 수 있습니다. 바이너리의 패스(path)는 관련 볼륨(volume)에 따라 달라집니다. 예시:
```
# Run with custom binary
BINARY=gaiafoo make localnet-start
```

## 멀티 노드, 리모트, 자동 테스트넷

다음 환경은 [네트워크 디렉터리](https://github.com/cosmos/cosmos-sdk/tree/develop/networks)에서 실행하셔야 합니다.

### Terraform 과 Ansible

자동 디플로이멘트(deployment)는 [Terraform](https://www.terraform.io/)를 이용해 AWS 서버를 만든 후 [Ansible](http://www.ansible.com/)을 이용해 해당 서버에서 테스트넷을 생성하고 관리하여 운영됩니다.

### 필수 사항

- [Terraform](https://www.terraform.io/downloads.html) 과 [Ansible](http://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)를 리눅스 머신에 설치.
- EC2 create 권한이 있는 [AWS API 토큰](https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html) 생성
- SSH 키 생성.

```
export AWS_ACCESS_KEY_ID="2345234jk2lh4234"
export AWS_SECRET_ACCESS_KEY="234jhkg234h52kh4g5khg34"
export TESTNET_NAME="remotenet"
export CLUSTER_NAME= "remotenetvalidators"
export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"
```

해당 명령은 `terraform` 과 `ansible`에서 이용됩니다..

### 리모트 네트워크 만들기

```
SERVERS=1 REGION_LIMIT=1 make validators-start
```

테스트넷 이름은 --chain-id에서 이용될 값이며, 클러스터 이름은 AWS 서버 관리 태그에서 이용될 값입니다. 코드는 각 존의 

The testnet name is what's going to be used in --chain-id, while the cluster name is the administrative tag in AWS for the servers. The code will create SERVERS amount of servers in each availability zone up to the number of REGION_LIMITs, starting at us-east-2. (us-east-1 is excluded.) The below BaSH script does the same, but sometimes it's more comfortable for input.

```
./new-testnet.sh "$TESTNET_NAME" "$CLUSTER_NAME" 1 1
```

### Quickly see the /status endpoint

```
make validators-status
```

### Delete servers

```
make validators-stop
```

### Logging

You can ship logs to Logz.io, an Elastic stack (Elastic search, Logstash and Kibana) service provider. You can set up your nodes to log there automatically. Create an account and get your API key from the notes on [this page](https://app.logz.io/#/dashboard/data-sources/Filebeat), then:

```
yum install systemd-devel || echo "This will only work on RHEL-based systems."
apt-get install libsystemd-dev || echo "This will only work on Debian-based systems."

go get github.com/mheese/journalbeat
ansible-playbook -i inventory/digital_ocean.py -l remotenet logzio.yml -e LOGZIO_TOKEN=ABCDEFGHIJKLMNOPQRSTUVWXYZ012345
```

### Monitoring

You can install the DataDog agent with:

```
make datadog-install
```
