// Package security provides helpers to protect sensitive values
// (such as provider API keys) before they are stored.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nuvora-core/nuvora/internal/config"
)

// secretKeyFile is the file that holds the AES key used to encrypt
// provider API keys at rest. It lives in the nuvora directory.
const secretKeyFile = "secret.key"

// EncryptAPIKey encrypts apiKey with AES-256-GCM and returns
// base64(nonce || ciphertext), which is safe to store in the database.
func EncryptAPIKey(apiKey string) (string, error) {
	key, err := loadOrCreateSecretKey()

	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)

	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends the ciphertext to the nonce, so the nonce travels
	// together with the data and is later needed for decryption.
	sealed := gcm.Seal(nonce, nonce, []byte(apiKey), nil)

	return base64.StdEncoding.EncodeToString(sealed), nil
}

// loadOrCreateSecretKey reads the AES key from the nuvora directory and
// generates a new random 32-byte key (0600) the first time it is needed.
func loadOrCreateSecretKey() ([]byte, error) {
	path := filepath.Join(config.GetNuvoraPath(), secretKeyFile)

	key, err := os.ReadFile(path)

	if err == nil {
		if len(key) != 32 {
			return nil, errors.New("invalid secret key length")
		}

		return key, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read secret key: %w", err)
	}

	key = make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate secret key: %w", err)
	}

	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, fmt.Errorf("write secret key: %w", err)
	}

	return key, nil
}
