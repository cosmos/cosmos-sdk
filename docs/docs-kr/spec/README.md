# 모듈 사양(Specifications)

이 디렉토리는 코스모스 SDK, 인터체인 표준(ICS, Interchain Standards) 등의 모듈들에 대한 사양을 설명합니다.

SDK 애플리케이션은 머클 스토어(Merkle store)에 상태를 보관합니다. 스토어에 대한 변경은 트랜잭션 중에 이루어지며 블록 전 또는 블록 후 이루어질 수 있습니다.

### SDK 사양:

- [Store](./store) - 애플리케이션의 샅애(state)를 보관하는 기본 머클 스토어.
- [Bech32](./other/bech32.md) - 코스모스 SDK 애플리케이션의 주소 포맷.

### 모듈 사양:

- [Auth](./auth) - 코스모스 계정과 트랜잭션의 표준 및 검증.
- [Bank](./bank) - 토큰 전송.
- [Governance](./governance) - 프로포절 및 투표.
- [Staking](./staking) - 지분증명 본딩, 위임, 등.
- [Slashing](./slashing) - 검증인 페널티 메커니즘.
- [Distribution](./distribution) - 스테이킹 보상과 스테이킹 토큰 규정 전파 등.
- [Inflation](./inflation) - 스테이킹 토큰 생성 규정.
- [IBC](./ibc) - Inter-Blockchain Communication (IBC) 프로토콜.

### 인터체인 표준

- [ICS30](./ics/ics-030-signed-messages.md) - 서명 메시지 표준.
-

기본 블록체인과 p2p 프로토콜에 대한 정보는 [여기](https://github.com/tendermint/tendermint/tree/develop/docs/spec)에서 확인하세요.
