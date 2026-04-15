# asmbase64

Assembly-optimized base64 backend for [github.com/lestrrat-go/jwx](https://github.com/lestrrat-go/jwx).

This module replaces jwx's default `encoding/base64` implementation with [github.com/segmentio/asm/base64](https://github.com/segmentio/asm), providing faster base64 encoding and decoding for JWK, JWS, and JWT operations.

## Installation

```
go get github.com/jwx-go/asmbase64/v4
```

## Usage

> **Process-global side effect.** Importing this package replaces the base64 backend for the entire process via `jwx.SetBase64Encoder` and `jwx.SetBase64Decoder`. Every JWS, JWE, JWK, and JWT operation in the binary — including code paths owned by other packages that also use jwx — will use the asm backend. Import from `main` or a top-level package, not from a leaf library, so the swap is visible at the binary boundary.

Import this package to activate the optimized base64 backend:

```go
import _ "github.com/jwx-go/asmbase64/v4"
```

This registers:

- **Base64 encoder**: assembly-optimized RawURL encoder via `jwx.SetBase64Encoder()`
- **Base64 decoder**: assembly-optimized decoder with automatic encoding detection via `jwx.SetBase64Decoder()`

### JWK round-trip

```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "encoding/json"

    _ "github.com/jwx-go/asmbase64"
    "github.com/lestrrat-go/jwx/v4/jwk"
)

key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
jwkKey, _ := jwk.Import[jwk.Key](key)
buf, _ := json.Marshal(jwkKey)
parsed, _ := jwk.ParseKey[jwk.Key](buf)
```

### JWS sign and verify

```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"

    _ "github.com/jwx-go/asmbase64"
    "github.com/lestrrat-go/jwx/v4/jwa"
    "github.com/lestrrat-go/jwx/v4/jws"
)

key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
signed, _ := jws.Sign(payload, jws.WithKey(jwa.ES256(), key))
verified, _ := jws.Verify(signed, jws.WithKey(jwa.ES256(), &key.PublicKey))
```

## License

MIT
