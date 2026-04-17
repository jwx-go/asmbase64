# AsmBase64 Extension for JWX

## Overview

This module (`github.com/jwx-go/asmbase64/v4`) provides an assembly-optimized base64 backend for `github.com/lestrrat-go/jwx`.

It replaces the default `encoding/base64` implementation with `github.com/segmentio/asm/base64`, which uses SIMD instructions for faster encoding/decoding. The replacement is transparent — importing this package is sufficient to activate the optimized backend for all jwx operations.

## Architecture

This module registers custom encoder and decoder implementations via `jwx.Settings(jwx.WithBase64Encoder(...), jwx.WithBase64Decoder(...))` in its `init()` function.

### Registration Points

| JWX Package | Registration Option | Purpose |
|-------------|---------------------|---------|
| `jwx` | `WithBase64Encoder()` (applied via `Settings`) | Replace default base64 encoder with asm-optimized RawURL encoder |
| `jwx` | `WithBase64Decoder()` (applied via `Settings`) | Replace default base64 decoder with asm-optimized auto-detecting decoder |

### Encoding Detection

The decoder automatically detects the base64 encoding variant (Std, RawStd, URL, RawURL) by inspecting the input for padding (`=`) and URL-unsafe characters (`+`, `/`).

## Build / Test

Requires `GOEXPERIMENT=jsonv2` (jwx v4 dependency):

```
GOEXPERIMENT=jsonv2 go test ./...
```

## Files

| File | Purpose |
|------|---------|
| `asmbase64.go` | Package doc, encoder/decoder types, `init()` registration, encoding detection |
| `asmbase64_test.go` | JWK and JWS round-trip tests |
| `appendencode_test.go` | Unit + fuzz coverage for `asmEncoder.AppendEncode` |
| `fuzz_test.go` | JWS sign/verify and JWK round-trip fuzz targets |

## Branch Policy

| Branch | Purpose |
|--------|---------|
| `v*` (e.g. `v4`) | Release tags only. NEVER commit directly to these branches. |
| `develop/v*` (e.g. `develop/v4`) | Active development. All feature branches merge here. |
| Feature branches | Branch from `develop/v*`, merge back via PR. |

- Tags are cut from `v*` branches.
- `v*` branches should never be directly worked on.
- Regular development happens on `develop/v*` and feature branches.
