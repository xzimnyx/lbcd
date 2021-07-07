// +build icu

package node

import (
	"math/rand"
	"testing"
)

func TestNormalizationICU(t *testing.T) {
	testNormalization(t, normalizeICU)
}

func BenchmarkNormalizeICU(b *testing.B) {
	benchmarkNormalizeICU(b, normalizeGo)
}
