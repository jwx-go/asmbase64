// Package asmbase64 provides an assembly-optimized base64 backend for jwx.
//
// Import this package for its side effects to replace the default
// encoding/base64 implementation with github.com/segmentio/asm/base64:
//
//	import _ "github.com/jwx-go/asmbase64/v4"
//
// # Process-global side effect
//
// This replacement is process-global. The init function calls
// jwx.Settings with jwx.WithBase64Encoder and jwx.WithBase64Decoder, which
// swap the backend used by every JWS, JWE, JWK, and JWT operation in the
// entire process, including code paths owned by other packages or
// third-party libraries that transitively depend on jwx. Import this
// package from main or a top-level package so the side effect is visible
// at the binary boundary; importing it from a leaf library silently
// changes behavior for every jwx caller in the program.
package asmbase64

import (
	"bytes"
	"fmt"
	"slices"

	jwx "github.com/lestrrat-go/jwx/v4"
	asmbase64 "github.com/segmentio/asm/base64"
)

func init() {
	jwx.Settings(
		jwx.WithBase64Encoder(asmEncoder{asmbase64.RawURLEncoding}),
		jwx.WithBase64Decoder(asmDecoder{}),
	)
}

type asmEncoder struct {
	*asmbase64.Encoding
}

func (e asmEncoder) AppendEncode(dst, src []byte) []byte {
	n := e.EncodedLen(len(src))
	dst = slices.Grow(dst, n)
	e.Encode(dst[len(dst):][:n], src)
	return dst[:len(dst)+n]
}

type asmDecoder struct{}

func (d asmDecoder) Decode(src []byte) ([]byte, error) {
	var enc *asmbase64.Encoding
	switch guess(src) {
	case encStd:
		enc = asmbase64.StdEncoding
	case encRawStd:
		enc = asmbase64.RawStdEncoding
	case encURL:
		enc = asmbase64.URLEncoding
	case encRawURL:
		enc = asmbase64.RawURLEncoding
	default:
		return nil, fmt.Errorf(`invalid encoding`)
	}

	dst := make([]byte, enc.DecodedLen(len(src)))
	n, err := enc.Decode(dst, src)
	if err != nil {
		return nil, fmt.Errorf(`failed to decode source: %w`, err)
	}
	return dst[:n], nil
}

const (
	encInvalid = iota
	encStd
	encURL
	encRawStd
	encRawURL
)

func guess(src []byte) int {
	isRaw := !bytes.HasSuffix(src, []byte{'='})
	isURL := !bytes.ContainsAny(src, "+/")
	switch {
	case isRaw && isURL:
		return encRawURL
	case isURL:
		return encURL
	case isRaw:
		return encRawStd
	default:
		return encStd
	}
}
