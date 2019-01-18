# 베이스코인 예시

이 항목에서는 간단한 베이스코인 블록체인을 시작하고, 계정 간 ``basecli`` 도구를 이용해서 트랜잭션을 보내고 어떻게 이 작업이 일어나는지 확인하겠습니다.

## 셋업 및 설치

이 튜토리얼을 진행하기 위해서는 고랭(golang)을 설치하셔야 합니다. 최신 코스모스 레포지토리와 고랭을 설치하는 방법이 있는 [코스모스 테스트넷 튜토리얼](https://cosmos.network/validators/tutorial) 문서를 참고하세요.

고랭을 설치하셨다면 다음 명령어를 실행하세요:

```
go get github.com/cosmos/cosmos-sdk
```

여기에서 `can't load package: package github.com/cosmos/cosmos-sdk: no Go files`라는 에러가 발생할 수 있으나 무시하셔도 좋습니다. 이제 디렉토리를 다음으로 바꿉니다:

```
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
```

이후 다음 명령어를 실행하세요 :

```
make tools // run make update_tools if you already had it installed
make get_vendor_deps
make install_examples
```
이후, `make install_examples`를 실행하세요. 해당 명령어는 `basecli`와 `basecoind` 바이너리를 생성합니다. 위 명령어가 어떻게 작동하는지 확인하기 위해서는 Makefile을 확인하시면 됩니다.

## basecli와 basecoind 사용하기

다음 명령어를 실행하여 버전을 확인합니다:

```
basecli version
basecoind version
```

명령어는 `0.17.1-5d18d5f`와 비슷한 값을 리턴합니다. 버전은 꾸준히 업데이트 되기 때문에 해당 버전보다 상위 버전이 표기되는 것은 문제가 되지 않습니다.


도움이 필요하시면 `basecli -h` 또는 `basecoind -h` 명령어를 실행하세요. 코드베이스가 변경되었을 수 있기 때문에 명령어에 관련한 정보를 꾸준히 확인하시는 것은 좋습니다.

이제 basecoind 데몬을 시작합니다. 다음 명령어를 입력하세요:

```
basecoind init
```

다음과 같은 값이 표기될 것입니다:

```
{
  "chain_id": "test-chain-z77iHG",
  "node_id": "e14c5056212b5736e201dd1d64c89246f3288129",
  "app_message": {
    "secret": "pluck life bracket worry guilt wink upgrade olive tilt output reform census member trouble around abandon"
  }
}
```

위 명령어는 `~/.basecoind folder`를 생성하며, 해당 폴더 안에는 config.toml, genesis.json, node_key.json, priv_validator.json 파일이 들어있습니다. 더 자세한 정보 또는 설정을 이해하시려면 해당 파일들의 내용을 확인하시는 것도 좋은 방법입니다.


## 키 생성하기

이제 priv_validator.son의 키를 gaiacli 키 매니저에 추가하는 단계입니다. 이번 단계에서는 프라이빗 키인 16자리 시드 단어와 비밀번호가 필요합니다. 위 명령어가 생성한 16자리 시드 단어는 `secret`에 있습니다.

이제 다음 명령어를 실행합니다:

```
basecli keys add alice --recover
```

위 명령어를 입력하시면 프로그램은 다음과 같은 요청을 합니다:

```
Enter a passphrase for your key:
Repeat the passphrase:
Enter your recovery seed phrase:
```

이제 로컬 환경에 저장된 'alice'라는 첫 키를 생성하셨습니다. 이 계정은 basecoind 검증인 노드를 운영하는 프라이빗 키와 연동되어있습니다. 위 절차를 진행하셨다면 생성하신 'alice' 키와 다른 키들을 보관하는 `~/.basecli` 폴더가 생성되었을 것입니다. 이제 alice키가 있으니 블록체인을 실행하시면 됩니다.

```
basecoind start
```

이제 빠른 속도로 블록이 생성되는 것을 목격하실 수 있습니다. 터미널에는 다수의 아웃풋이 표기될 것입니다.

이제 베이스코인의 '전송' 트랜잭션을 시용하기 위해서 더 많은 키를 생성해야합니다. 새로운 터미널을 실행하신 후 다음 명령어를 입력하셔서 두 개의 새로운 계정을 생성하세요. 비밀번호는 꼭 기억하셔야 합니다:

```
basecli keys add bob
basecli keys add charlie
```

키는 다음 명령어로 확인이 가능합니다:

```
basecli keys list
```

이제 alice, bob 그리고 charlie 계정이 표기될 것입니다.

```
NAME: 	ADDRESS:                                        PUBKEY:
alice   cosmos1khygs0qh7gz3p4m39u00mjhvgvc2dcpxhsuh5f	cosmospub1addwnpepq0w037u5g7y7lvdvsred2dehg90j84k0weyss5ynysf0nnnax74agrsxns6
bob     cosmos18se8tz6kwwfga6k2yjsu7n64e9z52nen29rhzz	cosmospub1addwnpepqwe97n8lryxrzvamrvjfj24jys3uzf8wndfvqa2l7mh5nsv4jrvdznvyeg6
charlie cosmos13wq5mklhn03ljpd4dkph5rflk5a3ssma2ag07q	cosmospub1addwnpepqdmtxv35rrmv2dvcr3yhfyxj7dzrd4z4rnhmclksq4g55a4wpl54clvx33l
```


## 'send' 트랜잭션

**Translator's note: 이 문서의 'send' 토큰을 한 계정에서 다른 계정으로 이동하는 행위를 뜻합니다. '트랜잭션'은 블록체인을 대상으로 발생하는 요청을 뜻합니다. 두 개념에 차이점을 참고하시기 바랍니다.**

이제 bob과 charlie에게 토큰을 전송할 시간입니다. 우선 Alice가 어떤 토큰을 보유하고 있는지 확인하시죠:


```
basecli account cosmos1khygs0qh7gz3p4m39u00mjhvgvc2dcpxhsuh5f
```

여기서 `cosmos1khygs0qh7gz3p4m39u00mjhvgvc2dcpxhsuh5f`는 `basecli keys list`를 실행할때 표기되는 alice의 주소입니다. 토큰 잔고 조회를 하시면 다수의 "mycoin"이 있는 것을 확인하실 수 있습니다. bob과 charlie 계정을 조회하시면 명령어는 실패합니다. 이는 bob과 charlie의 계정에 코인이 없기 때문에 계정이 블록체인에 기록되지 않았기 때문입니다. 이제 bob과 charlie에게 토큰을 보내면 됩니다.

다음 명령어는 alice의 토큰을 bob에게 보냅니다:

```
basecli send --from=alice --amount=10000mycoin --to=cosmos18se8tz6kwwfga6k2yjsu7n64e9z52nen29rhzz
--sequence=0 --chain-id=test-chain-AE4XQo
```

플래그 정보:
- `from`은 키에 할당한 이름입니다
- `mycoin`은 이 베이스코인 데모에서 사용되는 토큰의 명칭입니다. 이 명칭은 genesis.json 파일에서 설정됩니다
- `sequence`는 특정 계정이 발생한 트랜잭션의 총 수량입니다. 이번이 첫 트랜잭션이기 때문에 돌아오는 값은 0일 것입니다
- `chain-id`는 특정 블록체인/네트워크에 부여되는 고유 ID입니다. 이 ID는 텐더민트가 해당 블록체인을 식별할 수 있게 합니다. 관련 정보는 gaiad daemon이 실행되고 있는 터미널 아웃풋 중 '헤더' 섹션에서 확인하거나, `~/.basecoind/config/genesis.json` 디렉토리 내에 있는 genesis.json 파일에서 확인하실 수 있습니다

이제 bob의 계정을 확인하면 `10000 mycoin`의 잔고가 있는 것을 확인할 수 있습니다. bob의 계정을 확인하려면 다음 명령어를 실행하시면 됩니다 (주소를 확인하세요):

```
basecli account cosmos18se8tz6kwwfga6k2yjsu7n64e9z52nen29rhzz
```

이제 bob의 계정에서 charlie의 계정에 토큰을 보내봅시다. bob이 계정이 보유한 토큰보다 적은 수량을 보내는 것을 잊지 마세요, bob 계정의 토큰 보유량 이상의 트랜잭션은 실패합니다:

```
basecli send --from=bob --amount=5000mycoin --to=cosmos13wq5mklhn03ljpd4dkph5rflk5a3ssma2ag07q
--sequence=0 --chain-id=test-chain-AE4XQo
```

참고할 것은 ``--from`` 플래그를 이용해서 다른 계정의 토큰을 전송한다는 것입니다.

이제 bob의 계정에서 다시 alice의 계정으로 전송합니다:

```
basecli send --from=bob --amount=3000mycoin --to=cosmos1khygs0qh7gz3p4m39u00mjhvgvc2dcpxhsuh5f
--sequence=1 --chain-id=test-chain-AE4XQo
```

bob의 첫 트랜잭션을 `sequence 0`로 기록했기 때문에 이제 sequence의 값이 1입니다. 또한 터미널에서 리턴하는 ``hash`` 값을 참조하세요. 해당 값을 이용해 트랜잭션 기록을 조회할 수 있습니다. 명령어는 다음과 같습니다:

```
basecli tx <INSERT HASH HERE>
```

해당 명령어는 트랜잭션 해시값에 대한 정보를 리턴합니다. 여기에는 전송된 코인의 수량, 수신자 주소 그리고 트랜잭션이 발생한 블록 높이 정보 등이 포함됩니다.

이 것이 베이스코인의 기본적인 개념입니다!


## basecoind 블록체인과 basecli 데이터 리셋하기

**경고:** 다음 명령어를 실행하면 ``/.basecli``와 ``/.basecoind`` 디렉토리에 있는 (프라이빗 키를 포함한) 모든 정보를 삭제합니다. 베이스코인은 하나의 예시이기 때문에 큰 문제가 아니지만, 블록체인 시스템에서 프라이빗 키를 삭제하는 행동은 꼭 한번씩 더 확인하시며 진행하시기 바랍니다.

생성된 모든 파일을 삭제하고 작업 환경을 리셋 (튜토리얼을 다시 진행하거나 새로운 블록체인을 생성)하기 위해서는 다음 명령어를 실행하세요:

```
basecoind unsafe-reset-all
rm -rf ~/.basecoind
rm -rf ~/.basecli
```

## 베이스코인의 기술적 정보

이 항목은 베이스코인 애플리케이션 내에서 일어나는 작업들의 더 기술적인 정보를 다룹니다.

## 증거(Proof)

UI에서는 보이지 않을 수 있어도 모든 쿼리 값의 결과 값은 증거와 함께 전달됩니다. 증거 값은 해당 쿼리가 실제 블록체인에 포함되어있는지에 대한 머클 증거(Merkle proof) 값입니다. 상태(스테이트, state)의 머클 루트(Merkle root)는 최신 블록 헤더에 포함되어있습니다. ``basecli``는 밑단에서 상태(state) 값을 헤더 값과 대치하여 검증을 합니다. 또한 이 과정에서 올바른 검증인 세트가 블록 헤더를 서명했는지 검증합니다. 검증인 세트가 업데이트될 필요가 있을 경우, 보안이 확보되고 주요 변화가 없다는 가정 하에서, 검증인 세트를 변경하기도 합니다. 쿼리 진행에 다소 많은 시간이 소요되는 점에 대해 궁금하실 경우, 이런 많은 작업들이 클라이언트가 받는 정보가 중간의 풀노드에 의해 변형되지 않았다는 것을 검증한다는 것을 인지하시기 바랍니다.

## 계정과 트랜잭션(Accounts and Transactions)

도구들을 사용하는 방법을 더 이해하시고 싶으실 경우, 기반 데이터 형태(data structure)를 이해하는 것이 도움이 될 수도 있습니다. 이제 계정과 트랜잭션에 대해서 자세하게 알아보시다.

### 계정

베이스코인 상태(state)는 계정 세트(set of accounts)에 의해 구성됩니다. 각 계정에는 주소(address), 퍼블릭 키(public key), 다수 코인들의 잔고(coin denominations and balance) 그리고 리플레이 공격 방지 차원의 시퀀스 값이 포함됩니다. 이 형태의 계정은 이더리움의 계정 설계를 계승하였으며, 비트코인의 UTXO(Unspent Transaction Outputs)와는 다릅니다.

```
type BaseAccount struct {
  Address  sdk.Address   `json:"address"`
  Coins    sdk.Coins     `json:"coins"`
  PubKey   crypto.PubKey `json:"public_key"`
  Sequence int64         `json:"sequence"`
}
```

어카운트에 더 많은 값을 추가하는 것은 가능하며, 베이스코인 또한 값을 추가합니다. 베이스코인의 경우, `Name` 값을 추가하여 베이스 어카운트 디자인이 다양한 애플리케이션 사양에 따라 변경될 수 있는지 보여줍니다. 베이스코인은 위에 표기된 `auth.BaseAccount`를 기반으로 `Name` 항목을 추가합니다.

```
type AppAccount struct {
  auth.BaseAccount
  Name string `json:"name"`
}
```

계정 내에는 코인 잔고가 보관됩니다. 베이스코인은 다중 자산 암호화폐(multi-asset cryptocurrency)이기 때문에 각 계정에는 다양한 종류의 토큰이 보관될 수 있으며 어레이 형태로 저장됩니다.

```
type Coins []Coin

type Coin struct {
  Denom  string `json:"denom"`
  Amount int64  `json:"amount"`
}
```

블록체인에 더 새로운 코인을 추가하시려면 블록체인을 최초로 가동하기 전에 직접 ``/.basecoin/genesis.json`` 파일을 변경하셔야 합니다.

계정은 시리얼화(serialized)되어 키(``base/a/<address>``)의 머클 트리에 보관됩니다. 여기서 ``<address>``는 계정의 주소를 뜻합니다. 보통 계정의 주소는 퍼블릭 키의 ``sha256`` 해시 중 첫 20-바이트이나, `Tendermint crypto library <https://github.com/tendermint/tendermint/tree/master/crypto>`에 포함되어있는 다른 포맷도 이용이 가능합니다. 베이스코인이 이용하는 머클 트리는 balanced, bainary search tree이며, `IAVL 트리 <https://github.com/tendermint/iavl>`라는 명칭을 사용합니다.

### 트랜잭션(Transactions)

베이스코인은 다른 계정에게 토큰을 보내게 하는 `SendTx`를 하나의 트랜잭션 타입으로 구분합니다. `SendTx`는 인풋과 아웃풋의 리스트를 불러온 후, 인풋에 정의되어있는 계정의 토큰을 아웃풋에 정의되어있는 계정으로 전달합니다. `SendTx`의 구조는 다음과 같습니다:

```
type SendTx struct {
  Gas     int64      `json:"gas"`
  Fee     Coin       `json:"fee"`
  Inputs  []TxInput  `json:"inputs"`
  Outputs []TxOutput `json:"outputs"`
}

type TxInput struct {
  Address   []byte           `json:"address"`   // Hash of the PubKey
  Coins     Coins            `json:"coins"`     //
  Sequence  int              `json:"sequence"`  // Must be 1 greater than the last committed TxInput
  Signature []byte           `json:"signature"` // Depends on the PubKey type and the whole Tx
  PubKey    crypto.PubKey    `json:"pub_key"`   // Is present iff Sequence == 0
}

type TxOutput struct {
  Address []byte `json:"address"` // Hash of the PubKey
  Coins   Coins  `json:"coins"`   //
}
```

`SendTx`에 `Gas`와 `Fee` 항목을 참고할 필요가 있습니다. `Gas`는 트랜잭션이 발생시킬 수 있는 연산력을 제한하고, `Fee`는 수수료로 지불할 수량을 정의합니다. 이더리움의 `Gas`와 `GasPrice`에서 다른 점은 이더리움은 `Fee = Gas x GasPrice`이지만, 베이스코인은 `Gas`와 `Fee`는 독립적이며 `GasPrice`는 암시(implicit)된다는 점에서 차이가 있습니다.

베이스코인에서 `Fee`는 비트코인과 유사하게 검증인이 트랜잭션 순서를 정할때 사용하는 정보를 제공하기 위해 존재합니다. 그리고 `Gas`는 애플리케이션 플러그인이 트랜잭션의 실행을 컨트롤 하기 위해서 존재합니다. 현재 `Fee` 정보를 텐더민트 검증인에게 전달할 방법은 없으나, 추후 도입될 예정입니다. 그런 의미에서 이 버전의 베이스코인은 fee 와 gas를 완벽하게 도입하지 않으나, 계정간 트랜잭션을 발생시킬 수는 있습니다.


참고로 `PubKey`는 `Sequence == 0`일 경우에만 전송됩니다. 이후 트랜잭션에서는 머클 트리 내에 있는 계정 정보 내에 포함이 되어있으며, `Address`를 이용해 보내는 계정의 정보를 확인합니다. 이더리움의 경우 트랜잭션에 퍼블릭키를 포함하지 않지만, 이는 서명 정보에서 퍼블릭키 정보를 확인할 수 있는 다른 형태의 elliptic curve scheme을 이용하기 때문에 그렇습니다.

마지막으로, 다수의 인풋과 다수의 아웃풋을 이용하면 하나의 아토믹 트랜잭션(atomic transaction)을 이용해 다수의 계정 간 다수의 토큰을 전송할 수 잇게 합니다. 그렇기 때문에 `SendTx`는 탈중앙화 거래소의 기본적인 unit으로 이용될 수 있습니다. 여기서 주의해야할 점은, 인풋에 할당된 코인의 수량이 아웃풋에 할당된 코인의 수량과 동일해야한다는 것입니다(없는 돈을 생산할 수 없습니다). 또한 인풋을 제공하는 모든 계정이 서명을 했다는 것을 확인해야 합니다.
