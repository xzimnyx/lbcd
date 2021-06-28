// +build icu

package node

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizationICU(t *testing.T) {
	testNormalization(t, normalizeICU)
}

func BenchmarkNormalizeICU(b *testing.B) {
	benchmarkNormalizeICU(b, normalizeGo)
}
