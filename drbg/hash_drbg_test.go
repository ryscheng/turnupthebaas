package drbg

import (
	"testing"
)

func BenchmarkRandomUint32(b *testing.B) {
	drbg, _ := NewHashDrbg(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = drbg.RandomUint32()
	}
}

func BenchmarkRandomUint64(b *testing.B) {
	drbg, _ := NewHashDrbg(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = drbg.RandomUint64()
	}
}
