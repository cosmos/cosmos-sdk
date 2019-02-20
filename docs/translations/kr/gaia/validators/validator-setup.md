# 퍼블릭 테스트넷에서 밸리데이터 운영하기

::: tip
현재 테스트넷을 참가하는 방법은 [`testnet` repo](https://github.com/cosmos/testnets/tree/master/latest)에 있습니다. 최신 테스트넷에 대한 정보를 확인하시려면 해당 링크를 확인해주세요. 
:::

__Note__: 이 문서는 **퍼블릭 테스트넷** 검증인들을 위해서만 작성되었습니다.

밸리데이터 노드를 세팅하기 전, [풀노드 세팅](../join-testnet.md) 가이드를 먼저 확인해주세요.

## 밸리데이터란 무엇인가?

[밸리데이터](./overview.md)는 블록체인의 투표를 통해서 새로운 블록은 생성하는 역할을 합니다. 만약 특정 밸리데이터가 오프라인이 되거나, 같은 블록높이에서 중복 사이닝을 한 경우 해당 밸리데이터의 지분은 삭감(슬래싱, slashing) 됩니다. 노드를 DDOS 공격에서 보호하고 높은 접근성을 유지하기 위해서는 [센트리노드 아키텍쳐](./validator-faq.md#how-can-validators-protect-themselves-from-denial-of-service-attacks)에 대해서 읽어보세요.


::: danger 경고
코스모스 허브의 검증인이 되는 것을 검토하신다면, [보안에 대한 분석](./security.md)을 사전에 하시기를 바랍니다.
:::

만약 [풀노드](../join-testnet.md)를 이미 운영중이시다면, 다음 항목을 건너뛰셔도 좋습니다.

## 밸리데이터 생성하기

토큰 스테이킹을 통해 `cosmosvalconspub`로 새로운 밸리데이터를 생성할 수 있습니다. 본인의 밸리데이터 pubkey를 확인하기 위해서는 다음 명령어를 입력하세요:


```bash
gaiad tendermint show-validator
```

다음은 `gaiad gentx` 명령을 입력하세요:

::: warning 참고
보유하고 있는 `STAKE`이상을 이용하지 마십시오. 언제나 [Faucet](https://faucet.cosmos.network/)을 통해서 추가 `STAKE`를 받으실 수 있습니다.
:::

```bash
gaiacli tx staking create-validator \
  --amount=5STAKE \
  --pubkey=$(gaiad tendermint show-validator) \
  --moniker="choose a moniker" \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" 
```

__참고__: 커미션 값을 설정하실 때, `commission-max-change-rate`는 기존 `commission-rate`에서의 *퍼센트 포인트* 변경을 기준으로 측정됩니다. 예) 커미션이 1%에서 2%로 변경되는 것은 100% 상승되는 것이지만, 1%p 변경.

__참고__: `consensus_pubkey` 값이 지정되지 않은 경우, 기본적으로 `gaiad tendermint show-validator` 의 값으로 설정됩니다. `key_name`은 트랙잭션을 서명할때 이용되는 프라이빗키의 명칭입니다.

## 밸리데이터로써 제네시스 참가하기

__참고__: 이 문항은 제네시스 파일에 참가하려는 밸리데이터에게만 해당됩니다. 만약 검증을 하려는 체인이 이미 작동되고 있는 상태라면 이 항목을 건너뛰셔도 좋습니다.


밸리데이터로써 제네시스에 참가하고 싶으시다면 우선 본인(또는 위임자)가 stake를 보유하고 있다는 것을 증명해야 합니다. 스테이크를 검증인에게 본딩하는 하나 이상의 트랜잭션을 발생하신 후, 해당 트랜잭션을 제네시스 파일에 추가하시기 바랍니다.

우선 두가지의 케이스가 존재합니다:

- 경우 1: 본인 밸리데이터의 stake를 본딩(위임)한다.
- 경우 2: 타인(위임자)의 stake를 본딩한다.

### Case 1: 최초 위임이 밸리데이터 본인 주소에서 발생하는 경우

이런 경우에는 `gentx`를 생성하셔야 합니다:

```bash
gaiad gentx \
  --amount <amount_of_delegation> \
  --commission-rate <commission_rate> \
  --commission-max-rate <commission_max_rate> \
  --commission-max-change-rate <commission_max_change_rate> \
  --pubkey <consensus_pubkey> \
  --name <key_name>
```

__참고__: 이 명령어는 제네시스에서의 처리를 위해 `gentx`를 `~/.gaiad/config/gentx`에 저장합니다.

::: tip
명령어 플래그에 대한 정보는 `gaiad gentx --help`를 사용에 확인하십시오.
:::

`gentx`는 자체위임 정보가 포함된 JSON 파일입니다. 모든 제네시스 트랜잭셕은 `genesis coordinator`에 의하여 모아진 후 최초 `genesis.json`파일과 대치하여 검증합니다. 최초 `genesis.json`에는 계정 리스트와 각 계정이 보유하고 있는 코인 정보가 포함되어있습니다. 트랜잭션이 처리되었다면 해당 정보는 `genesis.json`의 `gentx` 항목에 머지(merge)됩니다.

### Case 2:  최초 위임이 위임자(delegator) 주소에서 발생하는 경우

이런 경우에는 위임자와 검증인의 서명이 둘다 필요합니다. 우선 서명이 되지 않은 `create-validator` 트랜잭션을 생성하신 후 `unsignedValTx`라는 파일에 저장하십시오:

```bash
gaiacli tx staking create-validator \
  --amount=5STAKE \
  --pubkey=$(gaiad tendermint show-validator) \
  --moniker="choose a moniker" \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --address-delegator="address of the delegator" \
  --generate-only \
  > unsignedValTx.json
```

이제 해당 `unsignedValTx`를 밸리데이터의 프라이빗 키를 이용해 서명합니다. 서명이된 아웃풋을 `signedValTx.json`이라는 파일에 저장합니다:

```bash
gaiacli tx sign unsignedValTx.json --from=<validator_key_name> > signedValTx.json
```

이제 이 파일을 위임자에게 전달하세요. 위임인은 다음 명령어를 실행하면 됩니다:

```bash
gaiacli tx sign signedValTx.json --from=<delegator_key_name> > gentx.json
```

이 파일은 제네시스 절차에서 필요하기 때문에 Case 1과 동일하게  `gentx.json`은 밸리데이터 머신의 `~/.gaiad/config/gentx` 폴더에 포함되어야 합니다 (Case 2 에서는 직접 해당 파을을 이동해야 합니다). 


### 제네시스 파일 복사, 제네시스 트랜잭션 처리하기

우선 `genesis.json`파일을 `gaiad`의 config 디렉토리로 가져옵니다.

Fetch the `genesis.json` file into `gaiad`'s config directory.

```bash
mkdir -p $HOME/.gaiad/config
curl https://raw.githubusercontent.com/cosmos/testnets/master/latest/genesis.json > $HOME/.gaiad/config/genesis.json
```

__참고:__ 이 항목에서는 최신 테스트넷 관련 정보가 있는 [테스트넷 repo](https://github.com/cosmos/testnets)의 `latest` 디렉토리를 사용합니다. 만약 다른 테스트넷에 연결하신다면 이용하시는 파일을 확인하시기 바랍니다.

이제 다른 제네시스 밸리데이터들의 제네시스 트랜잭션을 가져옵니다. 현재 밸리데이터들이 본인들의 제네시스 트랜잭션을 제공할 수 있는 리포지토리가 없는 상태이나, 추후 테스트넷에서 검증 후 추가될 예정입니다.

모든 제네시스 트랜잭션을 받으시고 `~/.gaiad/config/gentx`에 저장하셨다면 다음 명령어를 실행하십시오:

```bash
gaiad collect-gentxs
```

__참고:__ `gentx`에서 위임을 하는 계정에 스테이크(stake) 토큰이 있는 것을 확인하세요. 만약 해당 계정에 토큰이 없다면 `collect-gentx`가 실패하게 됩니다.

이전에 실행하신 명령어는 모든 제네시스 트랜잭션을 모으고 `genesis.json`을 파이널라이즈(finalize)합니다. 설정이 올바르게 되었는지 확인하기 위해서는 노드를 시작하십시오: 

```bash
gaiad start
```

## 검증인 설명 수정하기

검증인의 공개 설명 문구와 정보는 수정이 가능합니다. 이 정보는 위임자들이 어떤 검증인에게 위임을 할지 결정할때 이용될 수 있습니다. 각 플래그에 대해서 정보를 꼭 입력하시기 바랍니다. 만약 비어있는 항목이 있다면 해당 값은 빈 상태로 유지됩니다 (`--moniker`의 경우 머신 이름 값이 사용됩니다).

`--identity` 값은 Keybase 또는 UPort 같은 시스템을 이용해서 신분(identity)를 검증하는데 이용될 수 있습니다. Keybase를 사용하시는 경우 `--identity`는 [keybase.io](https://keybase.io) 계정으로 생성하신 16자리 string 값이 입력되어야 합니다. 이런 절차는 다수의 온라인 네트워크에서 본인의 신분을 증명하는데 이용될 수 있습니다. 또한 Keybase API를 이용해서 Keybase 아바타를 가져와 밸리데이터 프로파일에 이용하실 수 있습니다.

```bash
gaiacli tx staking edit-validator
  --moniker="choose a moniker" \
  --website="https://cosmos.network" \
  --identity=6A0D65E29A4CBC8E \
  --details="To infinity and beyond!" \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --commission-rate="0.10"
```

__참고__: `commission-rate` 값은 다음의 규칙을 따라야 합니다:

- 0 과 `commission-max-rate` 값의 사이
- 검증인의 `commission-max-change-rate` 값을 초과할 수 없습니다. `commission-max-change-rate`는 하루에 최대 커미션 값을 변경할 수 있는 한도입니다. 밸리데이터는 하루에 한번 `commission-max-change-rate`의 한도 안에서만 커미션을 변경할 수 있습니다.

## 밸리데이터 설명 확인하기

검증인의 정보는 다음 명령어로 확인이 가능합니다:

```bash
gaiacli query staking validator <account_cosmos>
```

## 밸리데이터 서명 정보 추적하기

특정 검증인의 과거 서명 정보를 확인하기 위해서는 `signing-info` 명령어를 이용하실 수 있습니다:

```bash
gaiacli query slashing signing-info <validator-pubkey>\
  --chain-id=<chain_id>
```

## 밸리데이터 석방(Unjail)하기

특정 검증인이 과도한 다운타임으로 '구속(jailed)' 상태로 전환되면 운영자의 계정에서 '석방(unjail)' 요청 트랜잭션을 전송해야만 다시 블록 생성 리워드를 받을 수 있습니다(각 존의 리워드 분배 정책에 따라 다를 수 있음).

```bash
gaiacli tx slashing unjail \
	--from=<key_name> \
	--chain-id=<chain_id>
```

## 밸리데이터 작동상태 확인

다음 명령어가 반응을 준다면 당신의 밸리데이터는 작동하고 있습니다:

```bash
gaiacli query tendermint-validator-set | grep "$(gaiad tendermint show-validator)"
```

코스모스 테스트넷의 경우 코스모스 [익스플로러](https://explorecosmos.network/validators)를 통해서 밸리데이터가 운영되고 있는지 확인하실 수 있습니다. `~/.gaiad/config/priv_validator.json` 파일의 `bech32` 인코딩이된 `address` 항목을 참고하세요.

::: warning 참고
검증인 세트에 포함되시기 원하신다면 100등 밸리데이터보다 보팅 파워(voting power)가 높아야 합니다.
:::

## 흔히 발생하는 문제들

### 문제 #1: 내 검증인의 보팅 파워가 0 입니다

밸리데이터가 자동 언본딩 되었습니다. `gaia-8000`의 경우, 100개 블록 중 50개의 블록에 투표하지 않을 경우 언본딩 됩니다. 블록은 대략 ~2초 마다 생성되기 때문에 ~100초 이상 비활성화 상태를 유지하는 밸리데이터는 언본딩 될 수 있습니다. 가장 흔한 이유는 `gaiad` 프로세스가 멈춘 경우입니다.

보팅 파워를 다시 밸리데이터에게 되돌리기 위해서, 우선 `gaiad`가 실행되고 있는지 확인하십시오. 만약 실행되고 있지 않은 경우 다음 명령어를 실행하십시오:

```bash
gaiad start
```

당신의 풀노드가 최신 블록높이에 싱크될때까지 기다리십시오. 이후, 다음 명령어를 실행하십시오. 참고로 `<cosmos>` 항목은 밸리데이터 계정의 주소이며, `<name>`은 밸리데이터 계정의 이름입니다. 해당 정보는 `gaiacli keys list` 명령어를 통해 확인하실 수 있습니다.

```bash
gaiacli tx slashing unjail <cosmos> --chain-id=<chain_id> --from=<from>
```

::: danger 경고
`gaiad`가 싱크되지 않은 상태에서 `unjail` 명령을 실행하실 경우, 검증인이 아직 '구속' 상태라는 메시지를 받게 됩니다.
:::

마지막으로 밸리데이터의 보팅파워가 복구 되었는지 확인하십시오.

```bash
gaiacli status
```

만약 보팅 파워가 예전보다 감소되었다면 다운타임에 대한 슬래싱이 이유일 수 있습니다.

### 문제 #2: `too many open files`때문에 `gaiad`가 강제 종료됩니다

리눅스가 각 프로세스당 열 수 있는는 파일 수는 최대 `1024`개입니다. `gaiad`는 1024개 이상의 열게될 수 있음으로 프로세스가 중단될 수 있습니다. 가장 간편한 해결책은 `ulimit -n 4096` (열 수 있는 최대 파일 수)명령어를 입력하고 프로세스를 `gaiad start`로 재시작하는 것입니다. 만약 `systemd` 또는 다른 프로세스 매니저로 `gaiad`를 실행하신다면 해당 레벨에서 몇가지 설정을 해야합니다. 문제 해결 샘플 `systemd` 파일은 다음과 같습니다:

```toml
# /etc/systemd/system/gaiad.service
[Unit]
Description=Cosmos Gaia Node
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu
ExecStart=/home/ubuntu/go/bin/gaiad start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```
