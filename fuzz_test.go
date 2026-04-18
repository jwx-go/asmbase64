package asmbase64_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jws"
	"github.com/stretchr/testify/require"

	// Import for side-effect registration of the asm base64 backend
	_ "github.com/jwx-go/asmbase64/v4"
)

func FuzzJWSSignAndVerify(f *testing.F) {
	f.Add([]byte("hello world"))
	f.Add([]byte(`{"key":"value"}`))
	f.Add([]byte(""))
	f.Add([]byte("The true sign of intelligence is not knowledge but imagination."))

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, payload []byte) {
		signed, err := jws.Sign(payload, jws.WithKey(jwa.ES256(), key))
		require.NoError(t, err)

		verified, err := jws.Verify(signed, jws.WithKey(jwa.ES256(), &key.PublicKey))
		require.NoError(t, err)
		require.Equal(t, payload, verified)
	})
}

func FuzzJWKRoundTrip(f *testing.F) {
	// Generate a seed: a valid EC P-256 JWK
	seedKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		f.Fatal(err)
	}
	jwkKey, err := jwk.Import[jwk.Key](seedKey)
	if err != nil {
		f.Fatal(err)
	}
	seedJSON, err := json.Marshal(jwkKey)
	if err != nil {
		f.Fatal(err)
	}

	f.Add(seedJSON)
	f.Add([]byte(""))
	f.Add([]byte("not-json"))

	f.Fuzz(func(_ *testing.T, data []byte) {
		parsed, err := jwk.ParseKeyAs[jwk.Key](data)
		if err != nil {
			return
		}

		buf, err := json.Marshal(parsed)
		if err != nil {
			return
		}

		_, _ = jwk.ParseKeyAs[jwk.Key](buf)
	})
}
