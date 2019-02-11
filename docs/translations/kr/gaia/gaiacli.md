# Gaia 클라이언트

## Gaia CLI

::: tip 참고
다음과 같은 에러 메시지가 나오는 경우:

```bash
Must specify these options: --chain-id  when --trust-node is false
```

라이트 클라이언트 증거를 검증할지 선택하셔야 합니다. 만약 쿼리를 요청하고 있는 노드를 신뢰할 수 있다면, `--trust-node=true`를 입력하시고, 그렇지 않다면 `--chain-id`를 입력하세요.
:::

`gaiacli`는 코스모스 테스트넷에서 이루어지는 트랜잭션과 계정을 관리하는 커맨드 라인 인터페이스입니다. 다음은 유용할 수 있는 `gaiacli` 명령어입니다:

`gaiacli`의 설정 파일은 `$HOME/.gaiacli/config/config.toml` 경로에 저장되며, 파일 수정 또는 `gaiacli config` 명령어를 통해 수정할 수 있습니다:

```bash
gaiacli config chain-id gaia-9004
```

명령어 사용에 관련한 정보는 help를 참고하세요: `gaiacli config --help`.

이 문서에는 유용한 `gaiacli` 명령어와 예시가 포함되어있습니다.

### 키(Keys)

#### 키 종류

키의 형태는 총 3개가 있습니다:

- `cosmos`
  - `gaiacli keyts add`로 생성되는 계정 키
  - 자금을 받는데 사용
  - 예시) `cosmos15h6vd5f0wqps26zjlwrc6chah08ryu4hzzdwhc`

* `cosmosvaloper`
  - 특정 검증인을 운영자와 연관하는데 사용됨
  - 스테이킹 명령 요청에 이용됨
  - 예시) `cosmosvaloper1carzvgq3e6y3z5kz5y6gxp3wpy3qdrv928vyah`

- `cosmospub`
  - `gaiacli keys add`로 생성되는 계정 키
  - 예시) `cosmospub1zcjduc3q7fu03jnlu2xpl75s2nkt7krm6grh4cc5aqth73v0zwmea25wj2hsqhlqzm`
- `cosmosvalconspub`
  - `gaiad init`로 새로운 노드가 생성될때 같이 생성되는 키.
  - `gaiad tendermint show-validator` 명령으로 키 값을 확인할 수 있음
  - 예시) `cosmosvalconspub1zcjduepq0ms2738680y72v44tfyqm3c9ppduku8fs6sr73fx7m666sjztznqzp2emf`

#### 키 생성하기

자금을 받거나, 트랜잭션을 전송하거나, 스테이킹을 하기 위해서는 프라이빗 키(`sk`)와 퍼블릭 키(`pk`) 쌍이 필요합니다.

새로운 _secp256k1_ 키를 생성하기 위해서는:

```bash
gaiacli keys add <account_name>
```

새로운 키를 생성하는 과정에서 나오는 _시트키(seed phrase)_ 는 안전하게 저장하시길 바랍니다. 시드키는 다음과 같은 명령을 이용해 잊어버린 퍼블릭/프라이빗 키를 복구하는데 이용됩니다:

```bash
gaiacli keys add --recover
```

이제 프라이빗 키를 확인하고 `<account_name>`을 찾으면 됩니다:

```bash
gaiacli keys show <account_name>
```

검증인 운영자 주소는 다음과 같이 확인하시고:

```shell
gaiacli keys show <account_name> --bech=val
```

관련 되어 있는 모든 키 목록은 다음 명령어로 찾으실 수 있습니다:

```bash
gaiacli keys list
```

본인 노드의 검증인 pubkey는 다음과 같이 확인할 수 있습니다:

```bash
gaiad tendermint show-validator
```

위 키는 텐더민트 사이닝 키이며, 위임 트랜잭션에서 이용되는 오퍼레이터 키가 아니라는 점을 참고하세요.

::: danger 경고
다수의 키에 동일한 passphrase를 이용하는 것을 추천하지 않습니다. 텐더민트 팀과 인터체인 재단은 자산 손실에 대한 책임을 지지 않습니다.
:::

#### 멀티시그 퍼블릭 키 생성하기

새로운 멀티시그 퍼블릭키를 생성하고 확인하시려면 다음과 같은 명령을 입력하세요:

```bash
gaiacli keys add --multisig=name1,name2,name3[...] --multisig-threshold=K new_key_name
```

여기서 `K`는 트랜잭션이 승인되기 위해서 필요한 최소의 키 개수입니다.

`--multisig` 플래그는 로컬 데이터베이스에 `new_key_name`으로 저장될 멀티시그 퍼블릭 키를 생성할때 사용되는 다수의 퍼블릭 키들의 값을 뜻합니다. `--multisig` 값에 포함될 키들은 로컬 데이터베이스에 존재하는 상태여야 합니다. `--nosort` 플래그가 정의된지 않은 경우, 멀티시그 조합에 필요한 키들이 입력되는 순서는 무관합니다. 예를 들어 다음 두 명령어는 두개의 동일한 멀티시그 퍼블릭 키를 생성합니다:

```bash
gaiacli keys add --multisig=foo,bar,baz --multisig-threshold=2 multisig_address
gaiacli keys add --multisig=baz,foo,bar --multisig-threshold=2 multisig_address
```

멀티시그 키의 주소는 다음과 같이 빠르게 생성하여 커맨드라인에 프린트할 수 있습니다:

```bash
gaiacli keys show --multisig-threshold K name1 name2 name3 [...]
```

멀티시그 트랜잭션를 생성, 사인, 전파하는 방법은 [멀티시그 트랜잭션](#멀티시그-트랜잭션) 항목을 참고하세요.

### 수수료와 가스

각 트랜잭션은 수수료(fee)를 지정하거나 가스 가격(gas price)을 지정할 수 있지만, 두 값을 함께 지정하는 것은 불가능합니다. 대다수의 유저는 지정된 비용만을 수수료로 사용하기 위해 수수료(fee)를 지정하는 방식으로 트랜잭션을 발생할 것으로 예상됩니다.

각 검증인은 최소 가스 가격(minimum gas price)를 (다수 토큰 사용 가능) 설정할 수 있으며 이 값을 기준으로 `CheckTx` 단계에서 특정 트랜잭션을 블록에 포함시킬지 확인합니다. `gasPrices >= minGasPrices`일때 검증인은 트랜잭션을 처리합니다. 참고로 트랜잭션 전파시 검증인이 요구하는 토큰 중 하나를 수수료 지불 토큰으로 사용하셔야 합니다.

__참고__: 위와 같은 메커니즘에서 일부 검증인은 멤풀에 있는 트랜잭션 중 높은 `gasPrice`의 트랜잭션을 우선적으로 처리할 수 있습니다. 그렇기 때문에 높은 수수료는 트랜잭션 처리 우선 순위를 높힐 수 있습니다.

예시)

```bash
gaiacli tx send ... --fees=100photino
```

또는

```bash
gaiacli tx send ... --gas-prices=0.000001stake
```

### 계정

#### 테스트 토큰 받기

토큰을 받기 가장 쉬운 곳은 [코스모스 테스트넷 faucet](https://faucetcosmos.network) 입니다. 만약 해당 faucet이 작동하지 않는 경우 [#cosmos-validators](https://riot.im/app/#/room/#cosmos-validators:matrix.org) 채팅 방에서 요청을 하실 수 있습니다. 해당 faucet은 스테이킹을 하려고 하시는 계정의 `cosmos` 주소를 입력하시면 됩니다.

#### 계정 잔고 조회하기

주소에 토큰을 받으신 후 잔고를 확인하시려면 다음 명령어를 입력하시면 됩니다:

```bash
gaiacli query account <account_cosmos>
```

::: warning 참고
계정의 토큰 잔고가 `0`인 계정을 조회하실 경우 다음과 같은 에러 메시지가 표시될 수 있습니다: `No account with address <account_cosmos> was found in the state.` 노드가 체인과 완벽하게 연동이 안된 상태에서 조회를 할 경우 동일한 에러가 발생할 수 있습니다.
:::

### 토큰 전송하기

한 계정에서 다른 계정으로 토큰/코인을 전송하기 위해서는 다음 명령어를 이용하시면 됩니다:

```bash
gaiacli tx send <destination_cosmos> 10faucetToken \
  --chain-id=<chain_id> \
  --from=<key_name> \
```

::: warning 참고
`--amount` 플래그는 다음과 같은 포맷을 사용합니다 `--amount=<수량|코인 이름>`.
:::

::: tip 참고
해당 트랜잭션이 사용하는 가스 값의 최대치를 설정하기 원하시면 `--gas` 플래그를 이용하세요. 만약 `--gas=auto`를 이용하시는 경우, 트랜잭션이 실행되기 전에 가스 서플라이가 자동으로 예측됩니다. 예측된 가스 값과 실제 트랜잭션이 일어나는 사이에 블록체인 상태가 변경될 수 있으며, 기존 예측 수량에서 값이 변경이 될 수 있다는 점을 유의하세요. 변경 값은 `--gas-adjustment` 플래그를 이용해 설정하실 수 있으며 기본 값은 1.0입니다.
:::

이제 토큰을 전송한 계정과 토큰을 받은 계정의 잔고를 확인합니다:

```bash
gaiacli query account <account_cosmos>
gaiacli query account <destination_cosmos>
```

특정 블록에서의 잔고를 확인하고 싶으시다면 `--block` 플래그를 사용하실 수 있습니다:

```bash
gaiacli query account <account_cosmos> --block=<block_height>
```

트랜잭션을 실제 전파하지 않고 시뮬레이션을 하시려면 명령어 뒤에 `--dry-run` 플래그를 추가하세요:

```bash
gaiacli tx send <destination_cosmosaccaddr> 10faucetToken \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --dry-run
```

또한 트랜잭션을 빌드한 후 해당 트랜잭션을 JSON 포맷으로 STDOUT에 프린트 하시기를 원하면 `--generate-only`를 명령어에 추가하시면 됩니다:

```bash
gaiacli tx send <destination_cosmosaccaddr> 10faucetToken \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --generate-only > unsignedSendTx.json
```

::: tip 참고
시뮬레이션은 퍼블릭 키를 필요로 하고 generate only는 키베이스를 사용하지 않기 때문에 시뮬레이션은 tx generate only 기능과 함께 사용될 수 없습니다.
:::

이제 `--generate-only`를 통해 프린트된 트랜잭션 파일을 서명하시려면 다음 명령어를 통해 키를 입력하시면 됩니다:

```bash
gaiacli tx sign \
  --chain-id=<chain_id> \
  --from=<key_name>
  unsignedSendTx.json > signedSendTx.json
```

트랜잭션의 서명을 검증하기 위해서는:

```bash
gaiacli tx sign --validate-signatures signedSendTx.json
```

서명된 트랜잭션을 노드로 전파하기 위해서는 JSON 파일을 다음 명령어를 통해 전달하면 됩니다:

```bash
gaiacli tx broadcast --node=<node> signedSendTx.json
```

### 트랜잭션 조회하기

#### 태그 매칭하기

트랜잭션 검색 명령을 이용하여 모든 트랜잭션에 추가되는 특정 `tags` 세트를 검색할 수 있습니다.

각 태그의 키-값 페어는 `<tag>:<value>` 형태로 이루어집니다. 더 상세한 검색을 원하실 경우 `&` 를 사용하여 태그를 추가할 수 있습니다.

`tag`를 이용한 트랜잭션 조회는 다음과 같이 합니다:

```bash
gaiacli query txs --tags='<tag>:<value>'
```

다수의 `tags`를 이용하실 경우:

```bash
gaiacli query txs --tags='<tag1>:<value1>&<tag2>:<value2>'
```

페이지네이션은 `page`와 `limit` 값으로 지원됩니다.

```bash
gaiacli query txs --tags='<tag>:<value>' --page=1 --limit=20
```

::: tip 참고

액션 태그는 관련 메시지의 `Type()` 명령이 응답하는 메시지 타입과 언제나 동일합니다.

각 SDK 모듈에 대한 `tags`는 여기에서 확인할 수 있습니다:

- [Common tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/types/tags.go#L57-L63)
- [Staking tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/staking/tags/tags.go#L8-L24)
- [Governance tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/gov/tags/tags.go#L8-L22)
- [Slashing tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/slashing/handler.go#L52)
- [Distribution tags](https://github.com/cosmos/cosmos-sdk/blob/develop/x/distribution/tags/tags.go#L8-L17)
- [Bank tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/bank/keeper.go#L193-L206)
:::

#### 트랜잭션 해시로 검색하기

다음과 같은 명령어를 이용하여 한 트랜잭션의 해시값을 이용해 조회를 할 수 있습니다:

```bash
gaiacli query tx [hash]
```

### 슬래싱

#### 언제일(Unjailing)

제일링 된 검증인을 언제일 하기 위해서는:

```bash
gaiacli tx slashing unjail --from <validator-operator-addr>
```

#### 서명 정보

특정 검증인의 서명 정보를 확인하기 위해서는:

```bash
gaiacli query slashing signing-info <validator-pubkey>
```

#### 슬래싱 파라미터 조회

현재 슬래싱 파라미터를 확인하기 위해서는:

```bash
gaiacli query slashing params
```

### 스테이킹

#### 검증인 세팅하기
검증인 후보가 되기 위한 가이드는 [검증인 세팅](../validators/validator-setup.md) 문서를 참고하세요.

#### 검증인에게 위임하기

메인넷에서는 `atom`을 특정 검증인에게 위임할 수 있습니다. 스테이킹에 참여하는 [위임인](/resources/delegators-faq)은 검증인 보상의 일부를 받을 수 있습니다. 관련 정보는 [코스모스 토큰 모델](https://github.com/cosmos/cosmos/raw/master/Cosmos_Token_Model.pdf)에서 확인하세요.

##### 검증인 조회하기

특정 체인의 모든 검증인 목록을 확인하기 위해서는 다음 명령을 실행하세요:

```bash
gaiacli query staking validators
```

특정 검증인에 대한 정보를 원하실 경우 다음 명령을 실행하세요:

```bash
gaiacli query staking validator <account_cosmosval>
```

#### 토큰 본딩하기

테스트넷의 경우 `atom`이 아닌 `steak`를 위임합니다. 특정 테스트넷 검증인에게 토큰을 본딩하기 위해서는:


```bash
gaiacli tx staking delegate \
  --amount=10steak \
  --validator=<validator> \
  --from=<key_name> \
  --chain-id=<chain_id>
```

`<validator>` 는 검증 대상 검증인의 운영자 주소입니다. 로컬 테스트넷을 운영하시는 경우, 다음 명령어로 관련 주소를 확인하실 수 있습니다:

```bash
gaiacli keys show [name] --bech val
```

여기에서`[name]`은 `gaiad`를 처음 설정하셨을때 정의한 키의 명칭입니다.

토큰이 본딩되고 있는 기간 동안에는 다른 본딩된 토큰과 함께 하나의 '풀'을 이룹니다. 검증인들과 위임인들은 해당 풀의 소유량에 비례하는 보상을 받게 됩니다.


::: tip 참고
보유하고 있는 `steak` 이상을 사용하지 마세요. `steak`가 더 필요한 경우 [Faucet](https://faucetcosmos.network/)에서 추가로 받으실 수 있습니다!
:::

##### 위임 조회

위임 요청을 검증인에게 전송한 경우, 관련 정보를 다음 명령을 통해 조회하실 수 있습니다:

```bash
gaiacli query staking delegation <delegator_addr> <validator_addr>

```

만약 특정 검증인에 대한 모든 위임 상태를 확인하고 싶으실 경우 다음 명령을 이용하세요:

```bash
gaiacli query staking delegation <delegator_addr>
```

과거 위임 기록에 대해서는 `--height` 플래그를 추가 하셔서 해당 블록 높이에 대한 기록을 조회하실 수 있습니다.

#### 토큰 언본딩 하기

만약 특정 검증인이 악의적인 행동을 했거나 또는 본인이 개인적인 이유로 일부 토큰을 언본딩을 워하는 경우 다음 명령어를 통해 토큰을 언본딩 할 수 있습니다. 언본딩은 정확한 수량인 `shares-amount`(예시, `12.1`) 또는 언본딩을 원하는 물량의 비율인 `shares-fraction`(예시, `0.25`) 값으로 표현될 수 있습니다.


```bash
gaiacli tx staking unbond \
  --validator=<account_cosmosval> \
  --shares-fraction=0.5 \
  --from=<key_name> \
  --chain-id=<chain_id>
```

언본딩은 언본딩 기간이 끝나는 대로 완료됩니다.

##### 언본딩 조회하기

언본딩 절차를 시작하신 후 관련 정보를 조회하는 방법은 다음과 같습니다:

```bash
gaiacli query staking unbonding-delegation <delegator_addr> <validator_addr>
```

또는 모든 언본딩 정보를 확인하고 싶으신 경우:

```bash
gaiacli query staking unbonding-delegations <account_cosmos>
```

추가적으로 특정 검증인으로 부터 언본딩하는 정보를 확인하고 싶으신 경우:

```bash
gaiacli query staking unbonding-delegations-from <account_cosmosval>
```

과거 언본딩 정보는 `--height` 플래그를 통해서 특정 블록 높이에 대한 언본딩 정보를 조회할 수 있습니다.

#### 재위임(Redelegate) 하기

재위임이란 본딩 되어있는 토큰을 한 검증인으로 부터 다른 검증인으로 옮기는 것입니다:

```bash
gaiacli tx staking redelegate \
  --addr-validator-source=<account_cosmosval> \
  --addr-validator-dest=<account_cosmosval> \
  --shares-fraction=50 \
  --from=<key_name> \
  --chain-id=<chain_id>
```

위 예시와 같이 재위임될 토큰의 수량은 특정 수량(`shares-amount`) 또는 일정 비율(`shares-fraction`)로 표현될 수 있습니다.

언본딩 기간이 지나면 재위임은 자동으로 완료됩니다.

##### 재위임 조회하기

재위임을 시작하신 후, 다음 명령을 통해서 관련 정보를 조회하실 수 있습니다:

```bash
gaiacli query staking redelegation <delegator_addr> <src_val_addr> <dst_val_addr>
```

모든 검증인에 대한 재위임을 확인하고 싶으신 경우:

```bash
gaiacli query staking redelegations <account_cosmos>
```

특정 검증인에 대한 재위임을 확인하고 싶으신 경우:

```bash
  gaiacli query staking redelegations-from <account_cosmosval>
```

과거 재위임에 대한 정보는 다른 트랜잭션과 동일하게 `--height` 플래그를 이용하여 특정 블록 높이에 대한 재위임 정보를 확인하실 수 있습니다.

#### 파라미터 조회

파라미터는 스테이킹의 하이-레벨 설정을 정의합니다. 현재 값은 다음 명령어를 통해서 조회할 수 있습니다:

```bash
gaiacli query staking params
```

위 명령어는 다음과 같은 정보를 표기합니다:

- 언본딩 기간
- 최대 검증인 수
- 스테이킹 코인 표기

해당 값은 거버넌스 절차의 `ParameterChange`(파라미터 변경) 프로포절을 통해서 변경됩니다.

#### 스테이킹 풀 조회하기

스티이킹 풀은 현재 상태(state)에 대한 다이내믹 파라미터(dynamic parameter)를 정의합니다. 관련 정보는 다음 명령을 통해 조회할 수 있습니다:

```bash
gaiacli query staking pool
```

`pool` 명령은 다음과 같은 정보에 대한 현재 값을 제공합니다:
- 본딩된 토큰 / 본딩 되어있지 않은 토큰
- 총 토큰 수량
- 연 인플레이션 비율과 가장 최근에 인플레이션이 변경된 블록 높이
- 가장 최근 기록된 bonded shares

##### 검증인 위임 조회하기

특정 검증인에 대한 모든 위임은 다음 명령으로 조회가 가능합니다:

```bash
  gaiacli query delegations-to <account_cosmosval>
```

### 거버넌스

거버넌스는 코스모스 허브의 유저가 소프트웨어 업그레이드, 메인넷 파라미터 또는 문서 형태의 프로포절 등에 대한 합의를 하는 절차입니다. 유저는 프로포절에 대한 투표를 함으로 이 절차에 참여할 수 있으며, 투표권은 메인넷 아톰 홀더들에게만 주어집니다.

다음은 투표 절차에 대한 정보입니다:

- 투표권은 본딩된 `Atom`을 소유한 유저에게만 주어지며, `본딩된 아톰 1개 = 1표` 기준으로 집계됩니다
- 투표권을 행사하지 않은 위임인들은 본인이 위임한 검증인들의 투표를 따르게 됩니다
- **검증인은 모든 프로포절에 투표를 해야합니다.** 프로포절에 투표를 하지 않는 검증인들은 부분적 슬래싱 페널티를 받게 됩니다
- 표는 각 프로포절의 투표 마감 시점(메인넷 기준 2주)에서 집계됩니다. 각 계정은 투표기간 중 표를 변경할 수 있으며(트랜잭션 수수료는 부과됩니다), 가장 마지막 표가 유효한 표로 집계됩니다
- 투표자들은 `Yes`, `No`, `NoWithVeto`와 `Abstain` 중에서 하나를 선택하여 투표할 수 있습니다
- 프로포절은 `(YesVote/(YesVotes+NoVotes+NoWithVetoVotes))>1/2` 또는 `(NoWithVetoVotes/(YesVotes+NoVotes+NoWithVetoVotes))<1/3`일 경우에만 통과하며, 이 외의 경우 부결됩니다.

거버넌스 절차에 대한 더 자세한 정보는 [거버넌스 모듈 스펙](./../spec/governance)을 확인하세요.

#### 거버넌스 프로포절 생성하기

새로운 거버넌스 프로포절을 생성하기 위해서는 프로포절 정보와 일정의 보증금을 예치해야 합니다:

- `title`: 프로포절 제목
- `description`: 프로포절 설명
- `type`: 프로포절 유형. _Text_ 형태여야 합니다 (_SoftwareUpgrade_ 와 _ParameterChange_ 는 아직 지원되지 않습니다).

```bash
gaiacli tx gov submit-proposal \
  --title=<title> \
  --description=<description> \
  --type=<Text/ParameterChange/SoftwareUpgrade> \
  --deposit=<40steak> \
  --from=<name> \
  --chain-id=<chain_id>
```

##### 프로포절 조회

프로포절이 생성된 후 관련 정보를 조회하는 방법은 다음과 같습니다:

```bash
gaiacli query gov proposal <proposal_id>
```

모든 프로포절에 대한 조회를 하기 위해서는:

```bash
gaiacli query gov proposals
```

프로포절을 `voter` 또는 `depositor`로 필터링 해서 조회할 수도 있습니다.

특정 거버넌스 프로포절의 제안자를 확인하기 위해서는:

```bash
gaiacli query gov proposer <proposal_id>
```

#### 보증금 추가하기

프로포절이 네트워크에 전파되기 위해서는 해당 프로포절의 보증금이 `minDeposit` 값 이상이어야 합니다 (현재 기본 값은 `10 steak`입니다). 만약 사전에 생성한 프로포절이 해당 기준을 충족하지 못하였다면 추후에 보증금을 추가 예치하여 활성화할 수 있습니다. 프로포절의 보증금이 최소 값을 도달하면 해당 프로포절의 투표는 활성화 됩니다:

```bash
gaiacli tx gov deposit <proposal_id> <200steak> \
  --from=<name> \
  --chain-id=<chain_id>
```

> _참고_: 기본 보증금 기준을 충족하지 못한 프로포절은 `MaxDepositPeriod`이 지나면 자동으로 삭제됩니다.

##### 보증금 조회하기

새로운 프로포절이 생성된 후, 해당 프로포절에 대한 보증금은 다음과 같이 조회할 수 있습니다:

```bash
gaiacli query gov deposits <proposal_id>
```

특정 주소에 대한 보증금은 다음과 같이 확인하실 수 있습니다:

```bash
gaiacli query gov deposit <proposal_id> <depositor_address>
```

#### 프로포절 투표하기

프로포절의 보증금이 `MinDeposit` 값에 도달하면 투표 기간이 시작됩니다. 본딩된 `Atom`을 보유한 홀더들은 각자 투표를 할 수 있습니다:


```bash
gaiacli tx gov vote <proposal_id> <Yes/No/NoWithVeto/Abstain> \
  --from=<name> \
  --chain-id=<chain_id>
```

##### 표 조회하기

특정 표와 관련한 정보를 조회하기 위해서는:

```bash
gaiacli query gov vote <proposal_id> <voter_address>
```
과거 프로포절에 대한 표 정보를 확인하기 위해서는:

```bash
gaiacli query gov votes <proposal_id>
```

#### 프로포절 결과 조회하기

특정 프로포절에 대한 결과를 확인하기 위해서는:

```bash
gaiacli query gov tally <proposal_id>
```

#### 거버넌스 파라미터 조회하기

현재 거버넌스 파라미터를 조회하기 위해서는:

```bash
gaiacli query gov param voting
gaiacli query gov param tallying
gaiacli query gov param deposit
```

### 스테이킹 리워드 분배

#### 리워드 분배 파라미터 조회

현재 리워드 분배 파라미터 값을 조회하기 위해서는:

```bash
gaiacli query distr params
```

#### 수령되지 않은 리워드를 받기

수령하지 않은 리워드를 수령하기 위해서는:

```bash
gaiacli query distr outstanding-rewards
```

#### 검증인 커미션 조회

특정 검증인의 커미션을 조회하기 위해서는:

```bash
gaiacli query distr commission <validator_address>
```

#### 검증인 슬래싱 조회

특정 검증인의 슬래싱 기록을 조회하기 위해서는:

```bash
gaiacli query distr slashes <validator_address> <start_height> <end_height>
```

#### 특정 검증인에서 수령되지 않은 리워드 조회

위임자의 특정 검증인에서 발생된 미수령 리워드를 조회하기 위해서는:

```bash
gaiacli query distr rewards <delegator_address> <validator_address>
```

#### 위임자의 수령 대기중인 모든 리워드 조회

위임자의 모든 수령 대기 리워드를 조회하기 위해서는:

```bash
gaiacli query distr rewards <delegator_address>
```

### 멀티시그 트랜잭션

멀티시그 트랜잭션을 서명하기 위해서는 다수의 프라이빗 기가 필요합니다. 그렇기 때문에 멀티시그 계정에서 트랜잭션을 생성하고 서명하기 위해서는 여러 인원간의 협동이 필요합니다. 멀티시그 키 보유자 누구나 트랜잭션을 생성할 수 있으며, 멀티시그 퍼블릭키를 생성하고 트랜잭션을 전파하기 위해서는 키 소유자 중 최소 한명이 다른 키 소유자들의 모든 퍼블릭 키를 로컬 데이터베이스에 보유해야합니다.

예를 들어 멀티시그 키가 `p1`, `p2`, `p3` 키로 이루어진다면, `p1` 키 보유자는 `p2`와 `p3`의 키가 있어야 멀티시그 계정의 퍼블릭 키를 생성할 수 있습니다.

```
gaiacli keys add \
  --pubkey=cosmospub1addwnpepqtd28uwa0yxtwal5223qqr5aqf5y57tc7kk7z8qd4zplrdlk5ez5kdnlrj4 \
  p2
 gaiacli keys add \
  --pubkey=cosmospub1addwnpepqgj04jpm9wrdml5qnss9kjxkmxzywuklnkj0g3a3f8l5wx9z4ennz84ym5t \
  p3
 gaiacli keys add \
  --multisig-threshold=2
  --multisig=p1,p2,p3
  p1p2p3
```

이제 새로운 멀티시그 키 `p1p2p3`이 보관되었으며 이 주소를 기반으로 멀티 트랜잭션이 서명됩니다:

```bash
gaiacli keys show --address p1p2p3
```

위 주소를 기반으로 멀티시그 트랜잭션을 생성하는 과정의 첫 단계는 다음과 같습니다:

```bash
gaiacli tx send cosmos1570v2fq3twt0f0x02vhxpuzc9jc4yl30q2qned 10stake \
  --from=<multisig_address> \
  --generate-only > unsignedTx.json
```

`unsignedTx.json` 파일은 서명되지 않은 트랜잭션을 JSON 형태로 보관합니다. 이제 `p1`은 본인의 프라이빗 키를 사용해 트랜잭션을 서명할 수 있습니다:

```bash
gaiacli tx sign \
  --multisig=<multisig_address> \
  --name=p1 \
  --output-document=p1signature.json \
  unsignedTx.json
```

서명이 생성된 후, `p1`은 `unsignedTx.json`과 `p1signature.json`을 `p2` 또는 `p3`에게 전다합니다. `p2`와 `p3`은 이를 기반으로 서명을 진행합니다:

```bash
gaiacli tx sign \
  --multisig=<multisig_address> \
  --name=p2 \
  --output-document=p2signature.json \
  unsignedTx.json
```

`p1p2p3`은 3명 중 2명의 서명을 필요로 하는 멀티시그 키입니다. 그렇기 때문에 `p1`이 서명한 트랜잭션에 하나의 프라이빗 키만 더해지면 트랜잭션이 유효합니다. 이제 다른 키 보유자들은 필요한 서명 파일을 결합하여 멀티시그 트랜잭션을 생성할 수 있습니다:

```bash
gaiacli tx multisign \
  unsignedTx.json \
  p1p2p3 \
  p1signature.json p2signature.json > signedTx.json
```

서명된 트랜잭션은 다음과 같은 명령을 실행하여 노드에 전파합니다:

```bash
gaiacli tx broadcast signedTx.json
```

## Shell 완료 스크립트

흔히 사용되는 `Bash`와 `Zsh` 같은 UNIX의 완료 스크립트(completion script)는 `completion` 명령어를 사용해 생성될 수 있습니다. 이 명령은 `gaiad`와 `gaiacli`에서 사용 가능합니다.

`Bash` 완료 스크립트를 생성하기 위해서는 다음 명령어를 실행하세요:

```bash
gaiad completion > gaiad_completion
gaiacli completion > gaiacli_completion
```

`Zsh` 완료 스크립트를 생성하기 위해서는 다음 명령어를 실행하세요:

```bash
gaiad completion --zsh > gaiad_completion
gaiacli completion --zsh > gaiacli_completion
```

::: tip 참고
대다수의 UNIX 시스템에서는 이런 스크립트를 `.bashrc` 또는 `.bash_profile`을 사용해 로딩할 수 있습니다:

```bash
echo '. gaiad_completion' >> ~/.bashrc
echo '. gaiacli_completion' >> ~/.bashrc
```

셸 자동 완성을 사용하시려면 사용하시는 OS의 매뉴얼을 참고하십시오.
:::
