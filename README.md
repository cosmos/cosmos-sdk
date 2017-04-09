

# crypto
`import "github.com/tendermint/go-crypto"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Subdirectories](#pkg-subdirectories)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [func CRandBytes(numBytes int) []byte](#CRandBytes)
* [func CRandHex(numDigits int) string](#CRandHex)
* [func CReader() io.Reader](#CReader)
* [func DecodeArmor(armorStr string) (blockType string, headers map[string]string, data []byte, err error)](#DecodeArmor)
* [func DecryptSymmetric(ciphertext []byte, secret []byte) (plaintext []byte, err error)](#DecryptSymmetric)
* [func EncodeArmor(blockType string, headers map[string]string, data []byte) string](#EncodeArmor)
* [func EncryptSymmetric(plaintext []byte, secret []byte) (ciphertext []byte)](#EncryptSymmetric)
* [func MixEntropy(seedBytes []byte)](#MixEntropy)
* [func Ripemd160(bytes []byte) []byte](#Ripemd160)
* [func Sha256(bytes []byte) []byte](#Sha256)
* [type PrivKey](#PrivKey)
  * [func PrivKeyFromBytes(privKeyBytes []byte) (privKey PrivKey, err error)](#PrivKeyFromBytes)
* [type PrivKeyEd25519](#PrivKeyEd25519)
  * [func GenPrivKeyEd25519() PrivKeyEd25519](#GenPrivKeyEd25519)
  * [func GenPrivKeyEd25519FromSecret(secret []byte) PrivKeyEd25519](#GenPrivKeyEd25519FromSecret)
  * [func (privKey PrivKeyEd25519) Bytes() []byte](#PrivKeyEd25519.Bytes)
  * [func (privKey PrivKeyEd25519) Equals(other PrivKey) bool](#PrivKeyEd25519.Equals)
  * [func (privKey PrivKeyEd25519) Generate(index int) PrivKeyEd25519](#PrivKeyEd25519.Generate)
  * [func (p PrivKeyEd25519) MarshalJSON() ([]byte, error)](#PrivKeyEd25519.MarshalJSON)
  * [func (privKey PrivKeyEd25519) PubKey() PubKey](#PrivKeyEd25519.PubKey)
  * [func (privKey PrivKeyEd25519) Sign(msg []byte) Signature](#PrivKeyEd25519.Sign)
  * [func (privKey PrivKeyEd25519) String() string](#PrivKeyEd25519.String)
  * [func (privKey PrivKeyEd25519) ToCurve25519() *[32]byte](#PrivKeyEd25519.ToCurve25519)
  * [func (p *PrivKeyEd25519) UnmarshalJSON(enc []byte) error](#PrivKeyEd25519.UnmarshalJSON)
* [type PrivKeyS](#PrivKeyS)
  * [func (p PrivKeyS) Empty() bool](#PrivKeyS.Empty)
  * [func (p PrivKeyS) MarshalJSON() ([]byte, error)](#PrivKeyS.MarshalJSON)
  * [func (p *PrivKeyS) UnmarshalJSON(data []byte) (err error)](#PrivKeyS.UnmarshalJSON)
* [type PrivKeySecp256k1](#PrivKeySecp256k1)
  * [func GenPrivKeySecp256k1() PrivKeySecp256k1](#GenPrivKeySecp256k1)
  * [func GenPrivKeySecp256k1FromSecret(secret []byte) PrivKeySecp256k1](#GenPrivKeySecp256k1FromSecret)
  * [func (privKey PrivKeySecp256k1) Bytes() []byte](#PrivKeySecp256k1.Bytes)
  * [func (privKey PrivKeySecp256k1) Equals(other PrivKey) bool](#PrivKeySecp256k1.Equals)
  * [func (p PrivKeySecp256k1) MarshalJSON() ([]byte, error)](#PrivKeySecp256k1.MarshalJSON)
  * [func (privKey PrivKeySecp256k1) PubKey() PubKey](#PrivKeySecp256k1.PubKey)
  * [func (privKey PrivKeySecp256k1) Sign(msg []byte) Signature](#PrivKeySecp256k1.Sign)
  * [func (privKey PrivKeySecp256k1) String() string](#PrivKeySecp256k1.String)
  * [func (p *PrivKeySecp256k1) UnmarshalJSON(enc []byte) error](#PrivKeySecp256k1.UnmarshalJSON)
* [type PubKey](#PubKey)
  * [func PubKeyFromBytes(pubKeyBytes []byte) (pubKey PubKey, err error)](#PubKeyFromBytes)
* [type PubKeyEd25519](#PubKeyEd25519)
  * [func (pubKey PubKeyEd25519) Address() []byte](#PubKeyEd25519.Address)
  * [func (pubKey PubKeyEd25519) Bytes() []byte](#PubKeyEd25519.Bytes)
  * [func (pubKey PubKeyEd25519) Equals(other PubKey) bool](#PubKeyEd25519.Equals)
  * [func (pubKey PubKeyEd25519) KeyString() string](#PubKeyEd25519.KeyString)
  * [func (p PubKeyEd25519) MarshalJSON() ([]byte, error)](#PubKeyEd25519.MarshalJSON)
  * [func (pubKey PubKeyEd25519) String() string](#PubKeyEd25519.String)
  * [func (pubKey PubKeyEd25519) ToCurve25519() *[32]byte](#PubKeyEd25519.ToCurve25519)
  * [func (p *PubKeyEd25519) UnmarshalJSON(enc []byte) error](#PubKeyEd25519.UnmarshalJSON)
  * [func (pubKey PubKeyEd25519) VerifyBytes(msg []byte, sig_ Signature) bool](#PubKeyEd25519.VerifyBytes)
* [type PubKeyS](#PubKeyS)
  * [func (p PubKeyS) Empty() bool](#PubKeyS.Empty)
  * [func (p PubKeyS) MarshalJSON() ([]byte, error)](#PubKeyS.MarshalJSON)
  * [func (p *PubKeyS) UnmarshalJSON(data []byte) (err error)](#PubKeyS.UnmarshalJSON)
* [type PubKeySecp256k1](#PubKeySecp256k1)
  * [func (pubKey PubKeySecp256k1) Address() []byte](#PubKeySecp256k1.Address)
  * [func (pubKey PubKeySecp256k1) Bytes() []byte](#PubKeySecp256k1.Bytes)
  * [func (pubKey PubKeySecp256k1) Equals(other PubKey) bool](#PubKeySecp256k1.Equals)
  * [func (pubKey PubKeySecp256k1) KeyString() string](#PubKeySecp256k1.KeyString)
  * [func (p PubKeySecp256k1) MarshalJSON() ([]byte, error)](#PubKeySecp256k1.MarshalJSON)
  * [func (pubKey PubKeySecp256k1) String() string](#PubKeySecp256k1.String)
  * [func (p *PubKeySecp256k1) UnmarshalJSON(enc []byte) error](#PubKeySecp256k1.UnmarshalJSON)
  * [func (pubKey PubKeySecp256k1) VerifyBytes(msg []byte, sig_ Signature) bool](#PubKeySecp256k1.VerifyBytes)
* [type Signature](#Signature)
  * [func SignatureFromBytes(sigBytes []byte) (sig Signature, err error)](#SignatureFromBytes)
* [type SignatureEd25519](#SignatureEd25519)
  * [func (sig SignatureEd25519) Bytes() []byte](#SignatureEd25519.Bytes)
  * [func (sig SignatureEd25519) Equals(other Signature) bool](#SignatureEd25519.Equals)
  * [func (sig SignatureEd25519) IsZero() bool](#SignatureEd25519.IsZero)
  * [func (p SignatureEd25519) MarshalJSON() ([]byte, error)](#SignatureEd25519.MarshalJSON)
  * [func (sig SignatureEd25519) String() string](#SignatureEd25519.String)
  * [func (p *SignatureEd25519) UnmarshalJSON(enc []byte) error](#SignatureEd25519.UnmarshalJSON)
* [type SignatureS](#SignatureS)
  * [func (p SignatureS) Empty() bool](#SignatureS.Empty)
  * [func (p SignatureS) MarshalJSON() ([]byte, error)](#SignatureS.MarshalJSON)
  * [func (p *SignatureS) UnmarshalJSON(data []byte) (err error)](#SignatureS.UnmarshalJSON)
* [type SignatureSecp256k1](#SignatureSecp256k1)
  * [func (sig SignatureSecp256k1) Bytes() []byte](#SignatureSecp256k1.Bytes)
  * [func (sig SignatureSecp256k1) Equals(other Signature) bool](#SignatureSecp256k1.Equals)
  * [func (sig SignatureSecp256k1) IsZero() bool](#SignatureSecp256k1.IsZero)
  * [func (p SignatureSecp256k1) MarshalJSON() ([]byte, error)](#SignatureSecp256k1.MarshalJSON)
  * [func (sig SignatureSecp256k1) String() string](#SignatureSecp256k1.String)
  * [func (p *SignatureSecp256k1) UnmarshalJSON(enc []byte) error](#SignatureSecp256k1.UnmarshalJSON)


#### <a name="pkg-files">Package files</a>
[armor.go](/src/github.com/tendermint/go-crypto/armor.go) [hash.go](/src/github.com/tendermint/go-crypto/hash.go) [priv_key.go](/src/github.com/tendermint/go-crypto/priv_key.go) [pub_key.go](/src/github.com/tendermint/go-crypto/pub_key.go) [random.go](/src/github.com/tendermint/go-crypto/random.go) [signature.go](/src/github.com/tendermint/go-crypto/signature.go) [symmetric.go](/src/github.com/tendermint/go-crypto/symmetric.go) 


## <a name="pkg-constants">Constants</a>
``` go
const (
    TypeEd25519   = byte(0x01)
    TypeSecp256k1 = byte(0x02)
    NameEd25519   = "ed25519"
    NameSecp256k1 = "secp256k1"
)
```
Types of implementations




## <a name="CRandBytes">func</a> [CRandBytes](/src/target/random.go?s=698:734#L28)
``` go
func CRandBytes(numBytes int) []byte
```
This uses the OS and the Seed(s).



## <a name="CRandHex">func</a> [CRandHex](/src/target/random.go?s=924:959#L38)
``` go
func CRandHex(numDigits int) string
```
RandHex(24) gives 96 bits of randomness, strong enough for most purposes.



## <a name="CReader">func</a> [CReader](/src/target/random.go?s=1078:1102#L43)
``` go
func CReader() io.Reader
```
Returns a crand.Reader mixed with user-supplied entropy



## <a name="DecodeArmor">func</a> [DecodeArmor](/src/target/armor.go?s=596:699#L18)
``` go
func DecodeArmor(armorStr string) (blockType string, headers map[string]string, data []byte, err error)
```


## <a name="DecryptSymmetric">func</a> [DecryptSymmetric](/src/target/symmetric.go?s=1048:1133#L23)
``` go
func DecryptSymmetric(ciphertext []byte, secret []byte) (plaintext []byte, err error)
```
secret must be 32 bytes long. Use something like Sha256(Bcrypt(passphrase))
The ciphertext is (secretbox.Overhead + 24) bytes longer than the plaintext.



## <a name="EncodeArmor">func</a> [EncodeArmor](/src/target/armor.go?s=125:206#L1)
``` go
func EncodeArmor(blockType string, headers map[string]string, data []byte) string
```


## <a name="EncryptSymmetric">func</a> [EncryptSymmetric](/src/target/symmetric.go?s=356:430#L6)
``` go
func EncryptSymmetric(plaintext []byte, secret []byte) (ciphertext []byte)
```
secret must be 32 bytes long. Use something like Sha256(Bcrypt(passphrase))
The ciphertext is (secretbox.Overhead + 24) bytes longer than the plaintext.
NOTE: call crypto.MixEntropy() first.



## <a name="MixEntropy">func</a> [MixEntropy](/src/target/random.go?s=407:440#L13)
``` go
func MixEntropy(seedBytes []byte)
```
Mix additional bytes of randomness, e.g. from hardware, user-input, etc.
It is OK to call it multiple times.  It does not diminish security.



## <a name="Ripemd160">func</a> [Ripemd160](/src/target/hash.go?s=185:220#L4)
``` go
func Ripemd160(bytes []byte) []byte
```


## <a name="Sha256">func</a> [Sha256](/src/target/hash.go?s=78:110#L1)
``` go
func Sha256(bytes []byte) []byte
```



## <a name="PrivKey">type</a> [PrivKey](/src/target/priv_key.go?s=326:435#L5)
``` go
type PrivKey interface {
    Bytes() []byte
    Sign(msg []byte) Signature
    PubKey() PubKey
    Equals(PrivKey) bool
}
```
PrivKey is part of PrivAccount and state.PrivValidator.







### <a name="PrivKeyFromBytes">func</a> [PrivKeyFromBytes](/src/target/priv_key.go?s=1302:1373#L50)
``` go
func PrivKeyFromBytes(privKeyBytes []byte) (privKey PrivKey, err error)
```




## <a name="PrivKeyEd25519">type</a> [PrivKeyEd25519](/src/target/priv_key.go?s=1502:1530#L58)
``` go
type PrivKeyEd25519 [64]byte
```
Implements PrivKey







### <a name="GenPrivKeyEd25519">func</a> [GenPrivKeyEd25519](/src/target/priv_key.go?s=3003:3042#L116)
``` go
func GenPrivKeyEd25519() PrivKeyEd25519
```

### <a name="GenPrivKeyEd25519FromSecret">func</a> [GenPrivKeyEd25519FromSecret](/src/target/priv_key.go?s=3290:3352#L125)
``` go
func GenPrivKeyEd25519FromSecret(secret []byte) PrivKeyEd25519
```
NOTE: secret should be the output of a KDF like bcrypt,
if it's derived from user input.





### <a name="PrivKeyEd25519.Bytes">func</a> (PrivKeyEd25519) [Bytes](/src/target/priv_key.go?s=1532:1576#L60)
``` go
func (privKey PrivKeyEd25519) Bytes() []byte
```



### <a name="PrivKeyEd25519.Equals">func</a> (PrivKeyEd25519) [Equals](/src/target/priv_key.go?s=1973:2029#L75)
``` go
func (privKey PrivKeyEd25519) Equals(other PrivKey) bool
```



### <a name="PrivKeyEd25519.Generate">func</a> (PrivKeyEd25519) [Generate](/src/target/priv_key.go?s=2761:2825#L106)
``` go
func (privKey PrivKeyEd25519) Generate(index int) PrivKeyEd25519
```
Deterministically generates new priv-key bytes from key.




### <a name="PrivKeyEd25519.MarshalJSON">func</a> (PrivKeyEd25519) [MarshalJSON](/src/target/priv_key.go?s=2156:2209#L83)
``` go
func (p PrivKeyEd25519) MarshalJSON() ([]byte, error)
```



### <a name="PrivKeyEd25519.PubKey">func</a> (PrivKeyEd25519) [PubKey](/src/target/priv_key.go?s=1826:1871#L70)
``` go
func (privKey PrivKeyEd25519) PubKey() PubKey
```



### <a name="PrivKeyEd25519.Sign">func</a> (PrivKeyEd25519) [Sign](/src/target/priv_key.go?s=1635:1691#L64)
``` go
func (privKey PrivKeyEd25519) Sign(msg []byte) Signature
```



### <a name="PrivKeyEd25519.String">func</a> (PrivKeyEd25519) [String](/src/target/priv_key.go?s=2613:2658#L101)
``` go
func (privKey PrivKeyEd25519) String() string
```



### <a name="PrivKeyEd25519.ToCurve25519">func</a> (PrivKeyEd25519) [ToCurve25519](/src/target/priv_key.go?s=2399:2453#L94)
``` go
func (privKey PrivKeyEd25519) ToCurve25519() *[32]byte
```



### <a name="PrivKeyEd25519.UnmarshalJSON">func</a> (\*PrivKeyEd25519) [UnmarshalJSON](/src/target/priv_key.go?s=2250:2306#L87)
``` go
func (p *PrivKeyEd25519) UnmarshalJSON(enc []byte) error
```



## <a name="PrivKeyS">type</a> [PrivKeyS](/src/target/priv_key.go?s=929:962#L30)
``` go
type PrivKeyS struct {
    PrivKey
}
```
PrivKeyS add json serialization to PrivKey










### <a name="PrivKeyS.Empty">func</a> (PrivKeyS) [Empty](/src/target/priv_key.go?s=1241:1271#L46)
``` go
func (p PrivKeyS) Empty() bool
```



### <a name="PrivKeyS.MarshalJSON">func</a> (PrivKeyS) [MarshalJSON](/src/target/priv_key.go?s=964:1011#L34)
``` go
func (p PrivKeyS) MarshalJSON() ([]byte, error)
```



### <a name="PrivKeyS.UnmarshalJSON">func</a> (\*PrivKeyS) [UnmarshalJSON](/src/target/priv_key.go?s=1057:1114#L38)
``` go
func (p *PrivKeyS) UnmarshalJSON(data []byte) (err error)
```



## <a name="PrivKeySecp256k1">type</a> [PrivKeySecp256k1](/src/target/priv_key.go?s=3635:3665#L136)
``` go
type PrivKeySecp256k1 [32]byte
```
Implements PrivKey







### <a name="GenPrivKeySecp256k1">func</a> [GenPrivKeySecp256k1](/src/target/priv_key.go?s=5071:5114#L194)
``` go
func GenPrivKeySecp256k1() PrivKeySecp256k1
```

### <a name="GenPrivKeySecp256k1FromSecret">func</a> [GenPrivKeySecp256k1FromSecret](/src/target/priv_key.go?s=5436:5502#L204)
``` go
func GenPrivKeySecp256k1FromSecret(secret []byte) PrivKeySecp256k1
```
NOTE: secret should be the output of a KDF like bcrypt,
if it's derived from user input.





### <a name="PrivKeySecp256k1.Bytes">func</a> (PrivKeySecp256k1) [Bytes](/src/target/priv_key.go?s=3667:3713#L138)
``` go
func (privKey PrivKeySecp256k1) Bytes() []byte
```



### <a name="PrivKeySecp256k1.Equals">func</a> (PrivKeySecp256k1) [Equals](/src/target/priv_key.go?s=4235:4293#L158)
``` go
func (privKey PrivKeySecp256k1) Equals(other PrivKey) bool
```



### <a name="PrivKeySecp256k1.MarshalJSON">func</a> (PrivKeySecp256k1) [MarshalJSON](/src/target/priv_key.go?s=4426:4481#L166)
``` go
func (p PrivKeySecp256k1) MarshalJSON() ([]byte, error)
```



### <a name="PrivKeySecp256k1.PubKey">func</a> (PrivKeySecp256k1) [PubKey](/src/target/priv_key.go?s=4032:4079#L151)
``` go
func (privKey PrivKeySecp256k1) PubKey() PubKey
```



### <a name="PrivKeySecp256k1.Sign">func</a> (PrivKeySecp256k1) [Sign](/src/target/priv_key.go?s=3772:3830#L142)
``` go
func (privKey PrivKeySecp256k1) Sign(msg []byte) Signature
```



### <a name="PrivKeySecp256k1.String">func</a> (PrivKeySecp256k1) [String](/src/target/priv_key.go?s=4673:4720#L177)
``` go
func (privKey PrivKeySecp256k1) String() string
```



### <a name="PrivKeySecp256k1.UnmarshalJSON">func</a> (\*PrivKeySecp256k1) [UnmarshalJSON](/src/target/priv_key.go?s=4522:4580#L170)
``` go
func (p *PrivKeySecp256k1) UnmarshalJSON(enc []byte) error
```



## <a name="PubKey">type</a> [PubKey](/src/target/pub_key.go?s=361:506#L7)
``` go
type PubKey interface {
    Address() []byte
    Bytes() []byte
    KeyString() string
    VerifyBytes(msg []byte, sig Signature) bool
    Equals(PubKey) bool
}
```
PubKey is part of Account and Validator.







### <a name="PubKeyFromBytes">func</a> [PubKeyFromBytes](/src/target/pub_key.go?s=1203:1270#L45)
``` go
func PubKeyFromBytes(pubKeyBytes []byte) (pubKey PubKey, err error)
```




## <a name="PubKeyEd25519">type</a> [PubKeyEd25519](/src/target/pub_key.go?s=1396:1423#L53)
``` go
type PubKeyEd25519 [32]byte
```
Implements PubKey










### <a name="PubKeyEd25519.Address">func</a> (PubKeyEd25519) [Address](/src/target/pub_key.go?s=1425:1469#L55)
``` go
func (pubKey PubKeyEd25519) Address() []byte
```



### <a name="PubKeyEd25519.Bytes">func</a> (PubKeyEd25519) [Bytes](/src/target/pub_key.go?s=1789:1831#L68)
``` go
func (pubKey PubKeyEd25519) Bytes() []byte
```



### <a name="PubKeyEd25519.Equals">func</a> (PubKeyEd25519) [Equals](/src/target/pub_key.go?s=3064:3117#L119)
``` go
func (pubKey PubKeyEd25519) Equals(other PubKey) bool
```



### <a name="PubKeyEd25519.KeyString">func</a> (PubKeyEd25519) [KeyString](/src/target/pub_key.go?s=2983:3029#L115)
``` go
func (pubKey PubKeyEd25519) KeyString() string
```
Must return the full bytes in hex.
Used for map keying, etc.




### <a name="PubKeyEd25519.MarshalJSON">func</a> (PubKeyEd25519) [MarshalJSON](/src/target/pub_key.go?s=2279:2331#L87)
``` go
func (p PubKeyEd25519) MarshalJSON() ([]byte, error)
```



### <a name="PubKeyEd25519.String">func</a> (PubKeyEd25519) [String](/src/target/pub_key.go?s=2823:2866#L109)
``` go
func (pubKey PubKeyEd25519) String() string
```



### <a name="PubKeyEd25519.ToCurve25519">func</a> (PubKeyEd25519) [ToCurve25519](/src/target/pub_key.go?s=2585:2637#L100)
``` go
func (pubKey PubKeyEd25519) ToCurve25519() *[32]byte
```
For use with golang/crypto/nacl/box
If error, returns nil.




### <a name="PubKeyEd25519.UnmarshalJSON">func</a> (\*PubKeyEd25519) [UnmarshalJSON](/src/target/pub_key.go?s=2372:2427#L91)
``` go
func (p *PubKeyEd25519) UnmarshalJSON(enc []byte) error
```



### <a name="PubKeyEd25519.VerifyBytes">func</a> (PubKeyEd25519) [VerifyBytes](/src/target/pub_key.go?s=1888:1960#L72)
``` go
func (pubKey PubKeyEd25519) VerifyBytes(msg []byte, sig_ Signature) bool
```



## <a name="PubKeyS">type</a> [PubKeyS](/src/target/pub_key.go?s=841:872#L25)
``` go
type PubKeyS struct {
    PubKey
}
```
PubKeyS add json serialization to PubKey










### <a name="PubKeyS.Empty">func</a> (PubKeyS) [Empty](/src/target/pub_key.go?s=1144:1173#L41)
``` go
func (p PubKeyS) Empty() bool
```



### <a name="PubKeyS.MarshalJSON">func</a> (PubKeyS) [MarshalJSON](/src/target/pub_key.go?s=874:920#L29)
``` go
func (p PubKeyS) MarshalJSON() ([]byte, error)
```



### <a name="PubKeyS.UnmarshalJSON">func</a> (\*PubKeyS) [UnmarshalJSON](/src/target/pub_key.go?s=964:1020#L33)
``` go
func (p *PubKeyS) UnmarshalJSON(data []byte) (err error)
```



## <a name="PubKeySecp256k1">type</a> [PubKeySecp256k1](/src/target/pub_key.go?s=3401:3430#L132)
``` go
type PubKeySecp256k1 [33]byte
```
Implements PubKey.
Compressed pubkey (just the x-cord),
prefixed with 0x02 or 0x03, depending on the y-cord.










### <a name="PubKeySecp256k1.Address">func</a> (PubKeySecp256k1) [Address](/src/target/pub_key.go?s=3497:3543#L135)
``` go
func (pubKey PubKeySecp256k1) Address() []byte
```
Implements Bitcoin style addresses: RIPEMD160(SHA256(pubkey))




### <a name="PubKeySecp256k1.Bytes">func</a> (PubKeySecp256k1) [Bytes](/src/target/pub_key.go?s=3774:3818#L145)
``` go
func (pubKey PubKeySecp256k1) Bytes() []byte
```



### <a name="PubKeySecp256k1.Equals">func</a> (PubKeySecp256k1) [Equals](/src/target/pub_key.go?s=4897:4952#L192)
``` go
func (pubKey PubKeySecp256k1) Equals(other PubKey) bool
```



### <a name="PubKeySecp256k1.KeyString">func</a> (PubKeySecp256k1) [KeyString](/src/target/pub_key.go?s=4814:4862#L188)
``` go
func (pubKey PubKeySecp256k1) KeyString() string
```
Must return the full bytes in hex.
Used for map keying, etc.




### <a name="PubKeySecp256k1.MarshalJSON">func</a> (PubKeySecp256k1) [MarshalJSON](/src/target/pub_key.go?s=4405:4459#L171)
``` go
func (p PubKeySecp256k1) MarshalJSON() ([]byte, error)
```



### <a name="PubKeySecp256k1.String">func</a> (PubKeySecp256k1) [String](/src/target/pub_key.go?s=4650:4695#L182)
``` go
func (pubKey PubKeySecp256k1) String() string
```



### <a name="PubKeySecp256k1.UnmarshalJSON">func</a> (\*PubKeySecp256k1) [UnmarshalJSON](/src/target/pub_key.go?s=4500:4557#L175)
``` go
func (p *PubKeySecp256k1) UnmarshalJSON(enc []byte) error
```



### <a name="PubKeySecp256k1.VerifyBytes">func</a> (PubKeySecp256k1) [VerifyBytes](/src/target/pub_key.go?s=3875:3949#L149)
``` go
func (pubKey PubKeySecp256k1) VerifyBytes(msg []byte, sig_ Signature) bool
```



## <a name="Signature">type</a> [Signature](/src/target/signature.go?s=204:304#L3)
``` go
type Signature interface {
    Bytes() []byte
    IsZero() bool
    String() string
    Equals(Signature) bool
}
```
Signature is a part of Txs and consensus Votes.







### <a name="SignatureFromBytes">func</a> [SignatureFromBytes](/src/target/signature.go?s=1031:1098#L40)
``` go
func SignatureFromBytes(sigBytes []byte) (sig Signature, err error)
```




## <a name="SignatureEd25519">type</a> [SignatureEd25519](/src/target/signature.go?s=1221:1251#L48)
``` go
type SignatureEd25519 [64]byte
```
Implements Signature










### <a name="SignatureEd25519.Bytes">func</a> (SignatureEd25519) [Bytes](/src/target/signature.go?s=1253:1295#L50)
``` go
func (sig SignatureEd25519) Bytes() []byte
```



### <a name="SignatureEd25519.Equals">func</a> (SignatureEd25519) [Equals](/src/target/signature.go?s=1520:1576#L58)
``` go
func (sig SignatureEd25519) Equals(other Signature) bool
```



### <a name="SignatureEd25519.IsZero">func</a> (SignatureEd25519) [IsZero](/src/target/signature.go?s=1352:1393#L54)
``` go
func (sig SignatureEd25519) IsZero() bool
```



### <a name="SignatureEd25519.MarshalJSON">func</a> (SignatureEd25519) [MarshalJSON](/src/target/signature.go?s=1701:1756#L66)
``` go
func (p SignatureEd25519) MarshalJSON() ([]byte, error)
```



### <a name="SignatureEd25519.String">func</a> (SignatureEd25519) [String](/src/target/signature.go?s=1420:1463#L56)
``` go
func (sig SignatureEd25519) String() string
```



### <a name="SignatureEd25519.UnmarshalJSON">func</a> (\*SignatureEd25519) [UnmarshalJSON](/src/target/signature.go?s=1797:1855#L70)
``` go
func (p *SignatureEd25519) UnmarshalJSON(enc []byte) error
```



## <a name="SignatureS">type</a> [SignatureS](/src/target/signature.go?s=648:685#L20)
``` go
type SignatureS struct {
    Signature
}
```
SignatureS add json serialization to Signature










### <a name="SignatureS.Empty">func</a> (SignatureS) [Empty](/src/target/signature.go?s=966:998#L36)
``` go
func (p SignatureS) Empty() bool
```



### <a name="SignatureS.MarshalJSON">func</a> (SignatureS) [MarshalJSON](/src/target/signature.go?s=687:736#L24)
``` go
func (p SignatureS) MarshalJSON() ([]byte, error)
```



### <a name="SignatureS.UnmarshalJSON">func</a> (\*SignatureS) [UnmarshalJSON](/src/target/signature.go?s=780:839#L28)
``` go
func (p *SignatureS) UnmarshalJSON(data []byte) (err error)
```



## <a name="SignatureSecp256k1">type</a> [SignatureSecp256k1](/src/target/signature.go?s=2013:2043#L80)
``` go
type SignatureSecp256k1 []byte
```
Implements Signature










### <a name="SignatureSecp256k1.Bytes">func</a> (SignatureSecp256k1) [Bytes](/src/target/signature.go?s=2045:2089#L82)
``` go
func (sig SignatureSecp256k1) Bytes() []byte
```



### <a name="SignatureSecp256k1.Equals">func</a> (SignatureSecp256k1) [Equals](/src/target/signature.go?s=2318:2376#L90)
``` go
func (sig SignatureSecp256k1) Equals(other Signature) bool
```



### <a name="SignatureSecp256k1.IsZero">func</a> (SignatureSecp256k1) [IsZero](/src/target/signature.go?s=2146:2189#L86)
``` go
func (sig SignatureSecp256k1) IsZero() bool
```



### <a name="SignatureSecp256k1.MarshalJSON">func</a> (SignatureSecp256k1) [MarshalJSON](/src/target/signature.go?s=2502:2559#L97)
``` go
func (p SignatureSecp256k1) MarshalJSON() ([]byte, error)
```



### <a name="SignatureSecp256k1.String">func</a> (SignatureSecp256k1) [String](/src/target/signature.go?s=2216:2261#L88)
``` go
func (sig SignatureSecp256k1) String() string
```



### <a name="SignatureSecp256k1.UnmarshalJSON">func</a> (\*SignatureSecp256k1) [UnmarshalJSON](/src/target/signature.go?s=2597:2657#L101)
``` go
func (p *SignatureSecp256k1) UnmarshalJSON(enc []byte) error
```







- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
