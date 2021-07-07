// +build use_icu_normalization

package node

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNormalizationICU(t *testing.T) {
	testNormalization(t, normalizeICU)
}

func BenchmarkNormalizeICU(b *testing.B) {
	benchmarkNormalize(b, normalizeICU)
}

func TestBlock760150(t *testing.T) {
	test := "Ꮖ-Ꮩ-Ꭺ-N--------Ꭺ-N-Ꮹ-Ꭼ-Ꮮ-Ꭺ-on-Instagram_-“Our-next-destination-is-East-and-Southeast-Asia--selfie--asia”"
	a := normalizeGo([]byte(test))
	b := normalizeICU([]byte(test))
	assert.Equal(t, a, b)
}