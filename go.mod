<<<<<<< HEAD
go 1.23.2
=======
go 1.26.3
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))

module github.com/cosmos/cosmos-sdk

require (
	cosmossdk.io/api v0.9.2
	cosmossdk.io/collections v1.3.1
	cosmossdk.io/core v0.11.3
	cosmossdk.io/depinject v1.2.1
	cosmossdk.io/errors v1.0.2
	cosmossdk.io/log v1.6.1
	cosmossdk.io/math v1.5.3
	cosmossdk.io/store v1.1.2
	cosmossdk.io/x/tx v0.14.0
	github.com/99designs/keyring v1.2.1
	github.com/bgentry/speakeasy v0.2.0
	github.com/bits-and-blooms/bitset v1.24.3
	github.com/chzyer/readline v1.5.1
	github.com/cockroachdb/apd/v2 v2.0.2
	github.com/cockroachdb/errors v1.12.0
	github.com/cometbft/cometbft v0.38.22
	github.com/cosmos/btcutil v1.0.5
	github.com/cosmos/cosmos-db v1.1.3
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/gogogateway v1.2.0
	github.com/cosmos/gogoproto v1.7.2
	github.com/cosmos/ledger-cosmos-go v1.0.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0
	github.com/golang/protobuf v1.5.4
	github.com/google/go-cmp v0.7.0
	github.com/google/gofuzz v1.2.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-metrics v0.5.4
	github.com/hashicorp/golang-lru v1.0.2
	github.com/hdevalence/ed25519consensus v0.2.0
	github.com/huandu/skiplist v1.2.1
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/jhump/protoreflect v1.17.0
	github.com/magiconair/properties v1.8.10
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.20
	github.com/mdp/qrterminal/v3 v3.2.1
<<<<<<< HEAD
	github.com/prometheus/client_golang v1.23.0
	github.com/prometheus/common v0.65.0
	github.com/rs/zerolog v1.34.0
=======
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/common v0.67.5
	github.com/rs/zerolog v1.35.1
	github.com/shirou/gopsutil/v4 v4.26.4
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/spf13/cast v1.10.0
	github.com/spf13/cobra v1.10.1
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/tendermint/go-amino v0.16.0
	go.uber.org/mock v0.6.0
<<<<<<< HEAD
	golang.org/x/crypto v0.41.0
	golang.org/x/sync v0.16.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.10
=======
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/crypto v0.51.0
	golang.org/x/sync v0.20.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260511170946-3700d4141b60
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	gotest.tools/v3 v3.5.2
	pgregory.net/rapid v1.2.0
	sigs.k8s.io/yaml v1.6.0
)

require (
<<<<<<< HEAD
	cosmossdk.io/schema v1.1.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/DataDog/datadog-go v3.2.0+incompatible // indirect
	github.com/DataDog/zstd v1.5.7 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.2 // indirect
	github.com/bytedance/sonic/loader v0.4.0 // indirect
=======
	cel.dev/expr v0.25.2 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.11.0 // indirect
	cloud.google.com/go/monitoring v1.29.0 // indirect
	cloud.google.com/go/storage v1.61.3 // indirect
	cosmossdk.io/schema v1.1.0 // indirect
	filippo.io/bigmod v0.1.1-0.20260103110540-f8a47775ebe5 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	filippo.io/keygen v0.0.0-20260114151900-8e2790ea4c5b // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/DataDog/datadog-go v3.2.0+incompatible // indirect
	github.com/DataDog/zstd v1.5.7 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.32.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.55.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.55.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.41.7 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.32.17 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.16 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.97.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.42.1 // indirect
	github.com/aws/smithy-go v1.25.1 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.6 // indirect
	github.com/bytedance/gopkg v0.1.4 // indirect
	github.com/bytedance/sonic v1.15.1 // indirect
	github.com/bytedance/sonic/loader v0.5.1 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
<<<<<<< HEAD
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240606204812-0bbfbd93a7ce // indirect
=======
	github.com/cloudwego/base64x v0.1.7 // indirect
	github.com/cncf/xds/go v0.0.0-20260202195803-dba9d589def2 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240816210425-c5d0cb0b6fc0 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/pebble v1.1.5 // indirect
	github.com/cockroachdb/redact v1.1.6 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cometbft/cometbft-db v0.14.1 // indirect
	github.com/cosmos/iavl v1.2.2 // indirect
	github.com/cosmos/ics23/go v0.11.0 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v4 v4.2.0 // indirect
	github.com/dgraph-io/ristretto v0.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.7.0 // indirect
<<<<<<< HEAD
	github.com/emicklei/dot v1.6.2 // indirect
	github.com/fatih/color v1.15.0 // indirect
=======
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/emicklei/dot v1.11.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.37.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/fatih/color v1.18.0 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
<<<<<<< HEAD
	github.com/getsentry/sentry-go v0.35.0 // indirect
=======
	github.com/getsentry/sentry-go v0.46.2 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/google/orderedcode v0.0.1 // indirect
<<<<<<< HEAD
=======
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.15 // indirect
	github.com/googleapis/gax-go/v2 v2.22.0 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
<<<<<<< HEAD
	github.com/hashicorp/go-plugin v1.6.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.12.0 // indirect
	github.com/linxGnu/grocksdb v1.8.14 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
=======
	github.com/hashicorp/go-plugin v1.7.0 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/huandu/go-assert v1.1.6 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ipfs/go-cid v0.5.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jbenet/go-temp-err-catcher v0.1.0 // indirect
	github.com/jhump/protoreflect/v2 v2.0.0-beta.2 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/koron/go-ssdp v0.0.6 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-flow-metrics v0.2.0 // indirect
	github.com/libp2p/go-libp2p v0.48.0 // indirect
	github.com/libp2p/go-libp2p-asn-util v0.4.1 // indirect
	github.com/libp2p/go-msgio v0.3.0 // indirect
	github.com/libp2p/go-netroute v0.4.0 // indirect
	github.com/libp2p/go-reuseport v0.4.0 // indirect
	github.com/libp2p/go-yamux/v5 v5.0.1 // indirect
	github.com/linxGnu/grocksdb v1.10.8 // indirect
	github.com/lufia/plan9stats v0.0.0-20260330125221-c963978e514e // indirect
	github.com/marcopolo/simnet v0.0.7 // indirect
	github.com/marten-seemann/tcp v0.0.0-20210406111302-dfbc87cc63fd // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	github.com/mikioh/tcpinfo v0.0.0-20190314235526-30a79bb1804b // indirect
	github.com/mikioh/tcpopt v0.0.0-20190314235656-172688c1accc // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/minio/highwayhash v1.0.4 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
<<<<<<< HEAD
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/petermattis/goid v0.0.0-20250813065127-a731cc31b4fe // indirect
	github.com/pkg/errors v0.9.1 // indirect
=======
	github.com/oasisprotocol/curve25519-voi v0.0.0-20251114093237-2ab5a27a1729 // indirect
	github.com/oklog/run v1.2.0 // indirect
	github.com/pascaldekloe/goe v0.1.1 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/petermattis/goid v0.0.0-20260330135022-df67b199bc81 // indirect
	github.com/pion/datachannel v1.5.10 // indirect
	github.com/pion/dtls/v3 v3.1.2 // indirect
	github.com/pion/ice/v4 v4.0.10 // indirect
	github.com/pion/interceptor v0.1.40 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.16 // indirect
	github.com/pion/rtp v1.8.19 // indirect
	github.com/pion/sctp v1.8.39 // indirect
	github.com/pion/sdp/v3 v3.0.18 // indirect
	github.com/pion/srtp/v3 v3.0.6 // indirect
	github.com/pion/stun/v3 v3.1.1 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/transport/v4 v4.0.1 // indirect
	github.com/pion/turn/v4 v4.0.2 // indirect
	github.com/pion/webrtc/v4 v4.1.2 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
<<<<<<< HEAD
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
=======
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.59.1 // indirect
	github.com/quic-go/webtransport-go v0.10.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.9 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
<<<<<<< HEAD
=======
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/zondax/golem v0.27.0 // indirect
	github.com/zondax/hid v0.9.2 // indirect
	github.com/zondax/ledger-go v1.0.1 // indirect
<<<<<<< HEAD
	go.etcd.io/bbolt v1.4.0-alpha.0.0.20240404170359-43604f3112c5 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.17.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
=======
	go.etcd.io/bbolt v1.4.3 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.43.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.68.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.43.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.19.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.19.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/fx v1.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/arch v0.26.0 // indirect
	golang.org/x/exp v0.0.0-20260508232706-74f9aab9d74a // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/telemetry v0.0.0-20260508192327-42602be52be6 // indirect
	golang.org/x/term v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.276.0 // indirect
	google.golang.org/genproto v0.0.0-20260511170946-3700d4141b60 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260511170946-3700d4141b60 // indirect
>>>>>>> c2739356e (chore: bump Go baseline to 1.26.3 and relax goconst linting (#26407))
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.6 // indirect
	rsc.io/qr v0.2.0 // indirect
)

// Here are the short-lived replace from the Cosmos SDK
// Replace here are pending PRs, or version to be tagged
// replace (
// 	<temporary replace>
// )

// Below are the long-lived replace of the Cosmos SDK
replace (
	// use cosmos fork of keyring
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	// dgrijalva/jwt-go is deprecated and doesn't receive security updates.
	// TODO: remove it: https://github.com/cosmos/cosmos-sdk/issues/13134
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.2
	// Fix upstream GHSA-h395-qcrw-5vmq and GHSA-3vp4-m3rf-835h vulnerabilities.
	// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.1
	// replace broken goleveldb
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)

retract (
	// incorrect tag for patch version
	v0.53.1

	// false start by tagging the wrong branch
	v0.50.0
	// revert fix https://github.com/cosmos/cosmos-sdk/pull/16331
	v0.46.12
	// subject to a bug in the group module and gov module migration
	[v0.46.5, v0.46.6]
	// subject to the dragonberry vulnerability
	// and/or the bank coin metadata migration issue
	[v0.46.0, v0.46.4]
	// subject to the dragonberry vulnerability
	[v0.45.0, v0.45.8]
	// do not use
	v0.43.0
)
