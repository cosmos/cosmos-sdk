# 최신 퍼블릭 테스트넷에 참가하기

::: tip 최신 테스트넷
최신 테스트넷에 대한 정보는 다음의 [테스트넷 리포](https://github.com/cosmos/testnets)를 참고하세요. 어떤 코스모스 SDK 버전과 제네시스 파일에 대한 정보가 포합되어있습니다.
:::

::: warning
**여기에서 더 진행하시기 전에 [gaia](./installation.md)가 꼭 설치되어있어야 합니다.**
:::

## 새로운 노드 세팅하기

> 참고: 과거 테스트넷에서 풀 노드를 운영하셨다면 이 항목은 건너뛰시고 [과거 테스트넷에서 업그레이드 하기](#upgrading-from-previous-testnet)를 참고하세요.

다음 절차는 새로운 풀노드를 처음부터 세팅하는 절차입니다.

우선 노드를 실행하고 필요한 config 파일을 생성합니다:

```bash
gaiad init --moniker <your_custom_moniker>
```

::: warning 참고
`--moniker`는 ASCII 캐릭터만을 지원합니다. Unicode 캐릭터를 이용하는 경우 노드 접근이 불가능할 수 있으니 참고하세요.
:::

`moniker`는 `~/.gaiad/config/config.toml` 파일을 통해 추후에 변경이 가능합니다:

```toml
# A custom human readable name for this node
moniker = "<your_custom_moniker>"
```

최소 수수료보다 낮은 트랜잭션을 거절하는 스팸 방지 메커니즘을 활성화 하시려면 `~/.gaiad/config/gaiad.toml` 파일을 변경하시면 됩니다:

```
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# Validators reject any tx from the mempool with less than the minimum fee per gas.
minimum_fees = ""
```


당신의 풀노드가 활성화 되었습니다! [제네시스와 시드](#genesis-seeds)로 넘어가주세요.

## 과거 테스트넷에서 업그레이드 하기

다음은 과거 테스트넷에서 운영을 했었던 풀노드가 최신 테스트넷으로 업그레이드를 하기 위한 절차입니다.

### 데이터 리셋

우선, 과거 파일을 삭제하고 데이터를 리셋합니다.

```bash
rm $HOME/.gaiad/config/addrbook.json $HOME/.gaiad/config/genesis.json
gaiad unsafe-reset-all
```

이제 `priv_validator.json`과 `config.toml`을 제외하고 노드가 초기화 되었습니다. 만약 해당 노드에 연결된적이 있는 센트리노드나 풀노드가 같이 업그레이드 되지 않았다면 연결이 실패할 수 있습니다.

::: danger 경고
각 노드가 **고유한** `priv_validator.json`을 보유하고 있는 것을 확인하세요. 오래된 노드의 `priv_validator.json`을 다수의 새로운 노드로 복사하지 마세요. 동일한 `priv_validator.json`을 보유한 두개 이상의 노드가 동시에 운영될 경우, **더블 사이닝**이 일어날 수 있습니다.
:::

### 소프트웨어 업그레이드

이제 소프트웨어를 업그레이드할 시간입니다:

```bash
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
git fetch --all && git checkout master
make update_tools install
```

::: tip
*참고*: 이 단계에서 문제가 있으시다면 최신 스테이블 GO 버전이 설치되어있는지 확인해주세요.
:::

위 예시에서는 가장 최신 스테이블 릴리즈가 있는 `master`를 사용합니다. 테스트넷마다 운용하는 릴리즈가 다를 경우가 있으니 [testnet repo](https://github.com/cosmos/testnets)를 확인하셔서 어떤 버전이 필요한지 확인하시고, [SDK 릴리즈 페이지](https://github.com/cosmos/cosmos-sdk/releases)에서 각 릴리즈에 대한 정보를 확인하세요.

이제 풀 노드가 깔끔하게 업그레이드 되었습니다!

## 제네시스와 시드

### 제네시스 파일 복사하기

테스트넷의 `genesis.json`파일을 `gaiad`의 config 디렉토리로 가져옵니다.

```bash
mkdir -p $HOME/.gaiad/config
curl https://raw.githubusercontent.com/cosmos/testnets/master/latest/genesis.json > $HOME/.gaiad/config/genesis.json
```

위 예시에서는 최신 테스트넷에 대한 정보가 포함되어있는 [테스트넷 repo]의 `latest`를 이용하는 것을 참고하세요. 만약 다른 테스트넷에 연결하신다면 해당 테스트넷의 파일을 가져오는 것인지 확인하세요.

설정이 올바르게 작동하는지 확인하기 위해서는 다음을 실행하세요:

```bash
gaiad start
```

### 시드 노드 추가하기

이제 노드가 다른 피어들을 찾는 방법을 알아야합니다. `$HOME/.gaiad/config/config.toml`에 안정적인 시드 노드들을 추가할 차례입니다. `testnets` repo에 각 테스트넷에 대한 시드 노드 링크가 포함되어있습니다. 만약 현재 운영되고 있는 테스트넷을 참가하고 싶으시다면 [여기](https://github.com/cosmos/testnets)에서 어떤 노드를 이용할지 확인하세요.

만약 해당 시드가 작동하지 않는다면, 추가적인 시드와 피어들을 [코스모스 익스플로러](https://explorer.cosmos.network/nodes)를 통해 확인하실 수 있습니다. `Full Nodes` 탭을 들어가셔서 프라이빗(`10.x.x.x`) 또는 로컬 IP 주소(https://en.wikipedia.org/wiki/Private_network)가 *아닌* 노드를 열어보세요. `Persistent Peer` 값에 연결 스트링(connection string)이 표기되어 있습니다. 가장 최적화된 결과를 위해서는 4-6을 이용하세요.

이 외에도 [밸리데이터 라이엇 채팅방](https://riot.im/app/#/room/#cosmos-validators:matrix.org)을 통해서 피어 요청을 할 수 있습니다.

시드와 피어에 대한 더 많은 정보를 원하시면 [여기](https://github.com/tendermint/tendermint/blob/develop/docs/tendermint-core/using-tendermint.md#peers)를 확인하세요.

## 풀노드 운영하기

다음 커맨드로 풀노드를 시작하세요:

```bash
gaiad start
```

모든 것이 잘 작동하고 있는지 확인하기 위해서는:

```bash
gaiacli status
```

네트워크 상태를 [코스모스 익스플로러](https://explorecosmos.network)를 통해 확인하세요. 현재 풀 노드가 현재 블록높이로 싱크되었을 경우, 익스플로러의 [풀 노드 리스트](https://explorecosmos.network/validators)에 표시가 될 것입니다. 익스플로러가 모든 노드에 연결하지는 않아 표시가 되지 않을 수 있다는 점 참고해주십시오.

## 상태 내보내기(Export State)

Gaia는 현재 애플리케이션의 상태를 JSON파일 형태로 내보낼 수 있습니다. 이런 데이터는 수동 분석과 새로운 네트워크의 제네시스 파일로 이용될 때 유용할 수 있습니다.

현재 상태를 내보내기 위해서는:

```bash
gaiad export > [filename].json
```

특정 블록 높이의 상태를 내보낼 수 있습니다(해당 블록 처리 후 상태):

```bash
gaiad export --height [height] > [filename].json
```

만약 해당 상태를 기반으로 새로운 네트워크를 시작하시려 한다면, `--for-zero-height` 플래그를 이용하셔서 내보내기를 실행해주세요:

```bash
gaiad export --height [height] --for-zero-height > [filename].json
```

## 밸리데이터 노드로 업그레이드 하기

이제 풀노드가 활성화 되었습니다! 다음은 무엇일까요?

이제는 해당 풀노드를 업그레이드 하여 코스모스 밸리데이터가 될 수 있습니다. 상위 100개 밸리데이터는 코스모스 허브의 블록 생성과 제안을 할 수 있습니다. 밸리데이터 노드로 업그레이드 하시기 위해서는 [밸리데이터 설정하기](./validators/validator-setup.md) 항목으로 넘어가주세요.
