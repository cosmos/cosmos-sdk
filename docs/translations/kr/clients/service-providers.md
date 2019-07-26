# 서비스 제공자(Service Providers)

'서비스 제공자'는 코스모스 SDK 기반 블록체인(코스모스 허브도 포함됩니다)과 교류하는 서비스를 엔드유저에게 제공하는 특정 인원/기관을 뜻합니다. 이 문서는 주로 토큰 인터랙션에 대한 정보를 다룹니다.

다음 항목은 [Light-Client](https://github.com/cosmos/cosmos-sdk/tree/develop/docs/light) 기능을 제공하려는 월렛 개발자들에게 해당하지 않습니다. 서비스 제공자는 엔드 유저와 블록체인을 이어주는 신뢰할 수 있는 기관/개인입니다.

## 보편적 아키텍처 설명

다음 세가지 항목을 고려해야 합니다:

- 풀 노드(Full-nodes): 블록체인과의 인터랙션. 
- REST 서버(Rest Server): HTTP 콜을 전달하는 역할.
- REST API: REST 서버의 활용 가능한 엔드포인트를 정의.

## 풀노드 운영하기

### 설치 및 설정

다음은 코스모스 허브의 풀노드를 설정하고 운용하는 방법입니다. 다른 코스모스 SDK 기반 블록체인 또한 비슷한 절차를 가집니다.

우선 [소프트웨어를 설치하세요](../getting-started/installation.md).

이후, [풀노드를 운영하세요](../getting-started/join-testnet.md).

### 커맨드 라인 인터페이스

다음은 풀노드를 이용할 수 있는 유용한 CLI 커맨드입니다.

#### 키페어 생성하기

새로운 키를 생성하기 위해서는 (기본적으로 secp256k1 엘립틱 커브 기반):

```bash
gaiacli keys add <your_key_name>
```

이후 해당 키페어에 대한 비밀번호(최소 8글지)를 생성할 것을 요청받습니다. 커맨드는 다음 4개 정보를 리턴합니다:


- `NAME`: 키 이름
- `ADDRESS`: 주소 (토큰 전송을 받을때 이용)
- `PUBKEY`: 퍼블릭 키 (검증인들이 사용합니다)
- `Seed phrase`: 12 단어 백업 시드키 **이 시드는 안전한 곳에 별도로 보관하셔야 합니다**. 이 시드키는 비밀번호를 잊었을 경우, 계정을 복구할때 사용됩니다.

다음 명령어를 통해서 사용 가능한 모든 키를 확인할 수 있습니다:

```bash
gaiacli keys list
```

#### 잔고 조회하기

해당 주소로 토큰을 받으셨다면 다음 명령어로 계정 잔고를 확인하실 수 있습니다:

```bash
gaiacli account <YOUR_ADDRESS>
```

*참고: 토큰이 0인 계정을 조회하실 경우 다음과 같은 에러 메시지가 표시됩니다: 'No account with address <YOUR_ADDRESS> was found in the state'. 해당 에러 메시지는 정상이며 앞으로 에러 메시지 개선이 들어갈 예정입니다.*

#### CLI로 코인 전송하기

다음은 CLI를 이용해 코인을 전송하는 명령어입니다:

```bash
gaiacli send --amount=10faucetToken --chain-id=<name_of_testnet_chain> --from=<key_name> --to=<destination_address>
```

플래그:
- `--amount`: `<value|coinName>` 포맷의 코인 이름/코인 수량입니다.
- `--chain-id`: 이 플래그는 특정 체인의 ID를 설정할 수 있게 합니다. 앞으로 테스트넷 체인과 메인넷 체인은 각자 다른 아이디를 보유하게 됩니다.
- `--from`: 전송하는 계정의 키 이름.
- `--to`: 받는 계정의 주소.

#### Help

이 외의 기능을 이용하시려면 다음 명령어를 사용하세요:

```bash
gaiacli 
```

사용 가능한 모든 명령어를 표기하며, 각 명령어 별로 `--help` 플래그를 사용하여 더 자세한 정보를 확인하실 수 있습니다.

## REST 서버 설정하기

REST 서버는 풀노드와 프론트엔드 사이의 중계역할을 합니다. REST 서버는 풀노드와 다른 머신에서도 운영이 가능합니다.

REST 서버를 시작하시려면: 

```bash
gaiacli advanced rest-server --node=<full_node_address:full_node_port>
```

플래그:
- `--trust-node`: 불리언 값. `true`일 경우, 라이트 클라이언트 검증 절차가 비활성화 됩니다. 만약 `false`일 경우, 절차가 활성화됩니다. 서비스 제공자의 경우 `true` 값을 이용하시면 됩니다. 기본 값은 `true`입니다.
- `--node`: 플노드의 주소와 포트를 입력하시면 됩니다. 만약 풀노드와 REST 서버가 동일한 머신에서 운영될 경우 주소 값은 `tcp://localhost:26657`로 설정하시면 됩니다.
- `--laddr`: REST 서버의 주소와 포트를 정하는 플래그입니다(기본 값 `1317`). 대다수의 경우에는 포트를 정하기 위해서 사용됩니다, 이 경우 주소는 "localhost"로 입력하시면 됩니다. 포맷은 <rest_server_address:port>입니다.


### 트랜잭션 수신 모니터링

추천하는 수신 트랜잭션을 모니터링하는 방식은 LCD의 다음 엔드포인트를 정기적으로 쿼리하는 것입니다:

[`/bank/balance/{account}`](https://cosmos.network/rpc/#/ICS20/get_bank_balances__address_)

## Rest API

REST API는 풀노드와 인터랙션이 가능한 모든 엔드포인트를 정의합니다. 다음 [링크](https://cosmos.network/rpc/)에서 확인이 가능합니다.

API는 엔드포인트의 카테고리에 따라 ICS 스탠다드로 나뉘어집니다. 예를 들어 [ICS20](https://cosmos.network/rpc/#/ICS20/)은 토큰 인터랙션 관련 API를 정의합니다.

서비스 제공자에게 더 많은 유연성을 제공하기 위해서 미서명 트랜잭션을 생성, [서명](https://cosmos.network/rpc/#/ICS20/post_tx_sign)과 [전달](https://cosmos.network/rpc/#/ICS20/post_tx_broadcast) 등의 다양한 API 엔드포인트가 제공됩니다. 이는 서비스 제공자가 자체 서명 메커니즘을 이용할 수 있게 합니다.

미서명 트랜잭션을 생성하기 위해서는 (예를 들어 [코인 전송](https://cosmos.network/rpc/#/ICS20/post_bank_accounts__address__transfers))을 생성하기 위해서는 `base_req` body에서 `generate_only` 플래그를 이용하셔야 합니다.