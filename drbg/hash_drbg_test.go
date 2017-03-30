package drbg

import (
	"bytes"
	"testing"
)

func TestOverlay(t *testing.T) {
	s, err := NewSeed()
	if err != nil {
		t.Fatal(err)
	}
	seed, err := s.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	initBuffer := make([]byte, 128)
	buffer := make([]byte, 128)
	err = Overlay(seed, buffer)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(initBuffer, buffer) == 0 {
		t.Fatal("Overlay failed to perturb buffer!")
	}

	_ = Overlay(seed, buffer)
	if bytes.Compare(initBuffer, buffer) != 0 {
		t.Fatal("Overlay applied twice should be identity!")
	}
}

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
