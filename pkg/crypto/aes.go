package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext format or corrupted data")
	ErrDecryptionFailed  = errors.New("decryption failed: authentication tag mismatch or invalid key")
)

// DeriveKey32 menurunkan kunci 32-byte (256-bit) dari sembarang passphrase menggunakan SHA-256
func DeriveKey32(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// Encrypt mengenkripsi plaintext menggunakan algoritma AES-256-GCM (Authenticated Encryption).
// Output berupa string hex yang menggabungkan [Nonce (12 bytes) + Ciphertext + Auth Tag (16 bytes)].
func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("AES-256 key must be exactly 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Buat nonce unik 12 bytes secara acak kriptografis
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	// Seal menambahkan ciphertext dan authentication tag ke nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt mendekripsi ciphertext hex string hasil enkripsi AES-256-GCM.
func Decrypt(ciphertextHex string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("AES-256 key must be exactly 32 bytes, got %d", len(key))
	}

	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex string: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, actualCiphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

