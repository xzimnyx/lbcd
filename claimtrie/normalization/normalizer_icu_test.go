// +build use_icu_normalization

package normalization

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
	"unicode/utf8"
)

func TestNormalizationICU(t *testing.T) {
	testNormalization(t, normalizeICU)
}

func BenchmarkNormalizeICU(b *testing.B) {
	benchmarkNormalize(b, normalizeICU)
}

func TestBlock760150(t *testing.T) {
	test, _ := hex.DecodeString("43efbfbd")
	assert.True(t, utf8.Valid(test))
	a := normalizeGo(test)
	b := normalizeICU(test)
	assert.Equal(t, a, b)

	test2 := "Ꮖ-Ꮩ-Ꭺ-N--------Ꭺ-N-Ꮹ-Ꭼ-Ꮮ-Ꭺ-on-Instagram_-“Our-next-destination-is-East-and-Southeast-Asia--selfie--asia”"
	a = normalizeGo([]byte(test2))
	b = normalizeICU([]byte(test2))
	assert.Equal(t, a, b)
}