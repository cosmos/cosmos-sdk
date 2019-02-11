# 레저(Ledger) 나노 하드웨어 지갑 지원

### 레저 시드를 이용한 계정 키 관리

이제 `gaiacli` 가 레저(Ledger) 시드를 통한 계정 키 관리를 지원합니다. 해당 기능을 이용하시기 위해서는 다음이 필요합니다:

- 이용하실 네트워크와 연결된 `gaiad` 인스턴스.
- 사용하시는 `gaiad` 인스턴스와 연동된 `gaiacli` 인스턴스.
- `ledger-cosmos` 앱이 설치된 레저 나노 기기.
  * 레저 기기에 코스모스 앱을 설치하기 위해서는 [`ledger-cosmos`](https://github.com/cosmos/ledger-cosmos/blob/master/docs/BUILD.md) 깃허브 레포지토리를 확인하세요.
  * 실전 운용 가능한 앱 버전은 추후 [레저 앱스토어](https://www.ledgerwallet.com/apps)를 통해 배포될 예정입니다.
  
> **참고:** 코스모스 키는 [BIP 44 Hierarchical Deterministic wallet 스펙](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)에 파생되었습니다. 관련 패스에 대한 정보는 [hd package](https://github.com/cosmos/cosmos-sdk/blob/develop/crypto/keys/hd/hdpath.go#L30)를 참고하세요.

코스모스 앱을 레저 기기에 성공적으로 설치하시고, `gaiacli` 에서 레저와 연결하는데 성공하셨다면 다음과 같이 레저 키를 생성하실 수 있습니다:

```bash
$ gaiacli keys add {{ .Key.Name }} --ledger
NAME:	          TYPE:	  ADDRESS:						                                  PUBKEY:
{{ .Key.Name }}	ledger	cosmos1aw64xxr80lwqqdk8u2xhlrkxqaxamkr3e2g943	cosmospub1addwnpepqvhs678gh9aqrjc2tg2vezw86csnvgzqq530ujkunt5tkuc7lhjkz5mj629
```

이 키는 레저가 연결되고 잠금 해제된 상태에서면 사용이 가능합니다. 해당 키를 이용해 코인을 전송하시려면 다음 명령을 실행하세요:


```bash
$ gaiacli tx send { .Destination.AccAddr } 10stake --from { .Key.Name } --chain-id=gaia-7000
```

레저 기기에서 해당 트랜잭션을 검토하신 후 서명이 되었다면 트랜잭션 결과를 레저 기기에서 확인하실 수 있습니다.
