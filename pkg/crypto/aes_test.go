package crypto_test

import (
	"strings"
	"testing"

	"symetra-lab-backend-v2/pkg/crypto"
)

func TestAES256GCM(t *testing.T) {
	passphrase := "my-super-secret-symetra-master-key"
	key := crypto.DeriveKey32(passphrase)

	t.Run("should successfully encrypt and decrypt plaintext", func(t *testing.T) {
		originalSecret := "shpk676b516461596a455257534a557973446969466f44455268686f6d727756"

		ciphertext, err := crypto.Encrypt(originalSecret, key)
		if err != nil {
			t.Fatalf("unexpected encryption error: %v", err)
		}

		if ciphertext == originalSecret {
			t.Fatalf("ciphertext must NOT match original secret")
		}

		decrypted, err := crypto.Decrypt(ciphertext, key)
		if err != nil {
			t.Fatalf("unexpected decryption error: %v", err)
		}

		if decrypted != originalSecret {
			t.Fatalf("expected decrypted %s, got %s", originalSecret, decrypted)
		}
	})

	t.Run("should produce different ciphertexts for same plaintext due to random nonce", func(t *testing.T) {
		secret := "partner-key-identical-value"
		cipher1, err1 := crypto.Encrypt(secret, key)
		cipher2, err2 := crypto.Encrypt(secret, key)

		if err1 != nil || err2 != nil {
			t.Fatalf("encryption failed: %v, %v", err1, err2)
		}

		if cipher1 == cipher2 {
			t.Fatalf("two encryptions of same secret must produce different ciphertexts")
		}

		// Both should still decrypt to same plaintext
		dec1, _ := crypto.Decrypt(cipher1, key)
		dec2, _ := crypto.Decrypt(cipher2, key)
		if dec1 != secret || dec2 != secret {
			t.Fatalf("both must decrypt correctly")
		}
	})

	t.Run("should fail to decrypt with wrong key", func(t *testing.T) {
		secret := "secret-data"
		ciphertext, err := crypto.Encrypt(secret, key)
		if err != nil {
			t.Fatalf("encryption error: %v", err)
		}

		wrongKey := crypto.DeriveKey32("different-passphrase-entirely")
		_, err = crypto.Decrypt(ciphertext, wrongKey)
		if err == nil {
			t.Fatalf("expected error when decrypting with wrong key, got nil")
		}
	})

	t.Run("should fail to decrypt corrupted or tampered ciphertext", func(t *testing.T) {
		secret := "secret-data"
		ciphertext, _ := crypto.Encrypt(secret, key)

		// Tamper with ciphertext by altering last characters
		tampered := ciphertext[:len(ciphertext)-4] + "ffff"
		_, err := crypto.Decrypt(tampered, key)
		if err == nil {
			t.Fatalf("expected error when decrypting tampered ciphertext, got nil")
		}

		// Test invalid hex
		_, err = crypto.Decrypt("not-a-valid-hex-string!!", key)
		if err == nil || !strings.Contains(err.Error(), "invalid hex") {
			t.Fatalf("expected invalid hex error, got %v", err)
		}
	})
}

