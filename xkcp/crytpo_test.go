package xkcp

import (
	"fmt"
	"testing"
)

func TestGetBlockCrypt(t *testing.T) {
	tests := []struct {
		name      string
		seed      string
		cryptType string
		wantNil   bool
	}{
		// Edge cases
		{"Empty seed and type", "", "", false}, // Should return default AES
		{"Empty seed", "", "aes-128", false},
		{"Invalid crypt type", "seed", "invalid", false}, // Should return default AES
		
		// Test all supported crypt types
		{"Null crypt", "test-seed", "null", true},
		{"SM4", "test-seed", "sm4", false},
		{"TEA", "test-seed", "tea", false},
		{"XOR", "test-seed", "xor", false}, 
		{"None", "test-seed", "none", false},
		{"AES-128", "test-seed", "aes-128", false},
		{"AES-192", "test-seed", "aes-192", false},
		{"Blowfish", "test-seed", "blowfish", false},
		{"Twofish", "test-seed", "twofish", false},
		{"Cast5", "test-seed", "cast5", false},
		{"3DES", "test-seed", "3des", false},
		{"XTEA", "test-seed", "xtea", false},
		{"Salsa20", "test-seed", "salsa20", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBlockCrypt(tt.seed, tt.cryptType)
			if (got == nil) != tt.wantNil {
				t.Errorf("GetBlockCrypt() got = %v, want nil = %v", got, tt.wantNil)
			}
		})
	}
}

func TestGetBlockCryptConsistency(t *testing.T) {
	seed := "test-seed"
	// Test that same inputs produce same outputs
	for _, cryptType := range []string{
		"sm4", "tea", "xor", "none", "aes-128", "aes-192",
		"blowfish", "twofish", "cast5", "3des", "xtea", "salsa20",
	} {
		t.Run(fmt.Sprintf("Consistency-%s", cryptType), func(t *testing.T) {
			first := GetBlockCrypt(seed, cryptType)
			second := GetBlockCrypt(seed, cryptType)

			if first == nil || second == nil {
				t.Fatal("unexpected nil result")
			}

			firstType := fmt.Sprintf("%T", first)
			secondType := fmt.Sprintf("%T", second)
			if firstType != secondType {
				t.Errorf("inconsistent types: got %v and %v", firstType, secondType)
			}
		})
	}
}

func TestGetBlockCryptDifferentSeeds(t *testing.T) {
	// Test that different seeds produce different ciphers
	cryptType := "aes-128"
	first := GetBlockCrypt("seed1", cryptType)
	second := GetBlockCrypt("seed2", cryptType)

	if first == nil || second == nil {
		t.Fatal("unexpected nil result")
	}

	// Verify different seeds produce different internal states
	// by comparing string representations
	if fmt.Sprintf("%v", first) == fmt.Sprintf("%v", second) {
		t.Error("different seeds produced identical ciphers")
	}
}
