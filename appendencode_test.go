package asmbase64

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"

	asm "github.com/segmentio/asm/base64"
	"github.com/stretchr/testify/require"
)

func newTestEncoder() asmEncoder {
	return asmEncoder{asm.RawURLEncoding}
}

func TestAppendEncode(t *testing.T) {
	t.Parallel()

	enc := newTestEncoder()

	for n := range 301 {
		src := make([]byte, n)
		_, err := rand.Read(src)
		require.NoError(t, err)

		want := base64.RawURLEncoding.EncodeToString(src)

		// (a) length invariant with empty dst
		got := enc.AppendEncode(nil, src)
		require.Len(t, got, enc.EncodedLen(len(src)),
			`n=%d: AppendEncode length should equal EncodedLen`, n)
		require.Equal(t, want, string(got), `n=%d: round-trip against stdlib`, n)

		// (b) prefix preservation with pre-populated dst
		prefix := []byte("prefix::")
		dst := make([]byte, len(prefix), len(prefix)+enc.EncodedLen(len(src))+64)
		copy(dst, prefix)
		got = enc.AppendEncode(dst, src)
		require.Equal(t, len(prefix)+enc.EncodedLen(len(src)), len(got),
			`n=%d: AppendEncode with prefix length`, n)
		require.True(t, bytes.HasPrefix(got, prefix), `n=%d: prefix must be preserved`, n)
		require.Equal(t, want, string(got[len(prefix):]), `n=%d: encoded suffix`, n)

		// (c) tail capacity must not be clobbered outside the returned slice
		dst = make([]byte, len(prefix), len(prefix)+enc.EncodedLen(len(src))+8)
		copy(dst, prefix)
		sentinel := dst[:cap(dst)][len(prefix)+enc.EncodedLen(len(src)):]
		for i := range sentinel {
			sentinel[i] = 0xAB
		}
		got = enc.AppendEncode(dst, src)
		tail := got[:cap(got)][len(got):]
		for i, b := range tail {
			require.Equal(t, byte(0xAB), b,
				`n=%d: tail byte %d at %d should not be clobbered`, n, b, i)
		}
	}
}

func TestAppendEncodeEmptySrc(t *testing.T) {
	t.Parallel()

	enc := newTestEncoder()

	got := enc.AppendEncode(nil, nil)
	require.Empty(t, got)

	dst := []byte("keep")
	got = enc.AppendEncode(dst, nil)
	require.Equal(t, "keep", string(got))
}

func TestNewEncoderStreaming(t *testing.T) {
	t.Parallel()

	enc := newTestEncoder()

	// Length 0..300 covers all boundary cases around the 3-byte grouping
	// that base64 encoders buffer internally. Each length is encoded in
	// three granularities — one full Write, byte-at-a-time, and a half+
	// half split — to catch any assumptions about Write() call
	// boundaries.
	for n := range 301 {
		src := make([]byte, n)
		_, err := rand.Read(src)
		require.NoError(t, err)

		want := base64.RawURLEncoding.EncodeToString(src)

		// Full buffer in one Write.
		var full bytes.Buffer
		wc := enc.NewEncoder(&full)
		_, err = wc.Write(src)
		require.NoError(t, err, `n=%d: Write`, n)
		require.NoError(t, wc.Close(), `n=%d: Close`, n)
		require.Equal(t, want, full.String(), `n=%d: single-write output`, n)

		// Byte-by-byte — forces the encoder to buffer partial groups.
		var byByte bytes.Buffer
		wc = enc.NewEncoder(&byByte)
		for _, b := range src {
			_, err = wc.Write([]byte{b})
			require.NoError(t, err, `n=%d: byte-wise Write`, n)
		}
		require.NoError(t, wc.Close(), `n=%d: byte-wise Close`, n)
		require.Equal(t, want, byByte.String(), `n=%d: byte-wise output`, n)

		// Half-and-half — exercises group-spanning Writes.
		if n >= 2 {
			var split bytes.Buffer
			wc = enc.NewEncoder(&split)
			_, err = wc.Write(src[:n/2])
			require.NoError(t, err, `n=%d: split Write(1)`, n)
			_, err = wc.Write(src[n/2:])
			require.NoError(t, err, `n=%d: split Write(2)`, n)
			require.NoError(t, wc.Close(), `n=%d: split Close`, n)
			require.Equal(t, want, split.String(), `n=%d: split output`, n)
		}
	}
}

func FuzzAppendEncode(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add([]byte("hello"))
	f.Add(bytes.Repeat([]byte{0xFF}, 64))

	enc := newTestEncoder()

	f.Fuzz(func(t *testing.T, src []byte) {
		want := base64.RawURLEncoding.EncodeToString(src)

		got := enc.AppendEncode(nil, src)
		require.Equal(t, want, string(got))

		prefix := []byte("p:")
		dst := make([]byte, len(prefix), len(prefix)+enc.EncodedLen(len(src))+4)
		copy(dst, prefix)
		got = enc.AppendEncode(dst, src)
		require.True(t, bytes.HasPrefix(got, prefix))
		require.Equal(t, want, string(got[len(prefix):]))
	})
}
