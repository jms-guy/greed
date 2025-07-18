package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

type Encryptor struct{}

func NewEncryptor() *Encryptor {
	return &Encryptor{}
}

// EncryptAccessToken securely encrypts the plaintext token using AES-GCM.
// The returned string is base64-encoded and safe to store.
func (e *Encryptor) EncryptAccessToken(plaintext []byte, keyString string) (string, error) {

    key, err := hex.DecodeString(keyString) // 32 bytes for AES-256
    if err != nil {
        return "", fmt.Errorf("error decoding key string to bytes: %w", err)
    }
    c, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("error creating new cipher: %w", err)
    }

    gcm, err := cipher.NewGCM(c)
    if err != nil {
        return "", fmt.Errorf("error generating new GCM: %w", err)
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", fmt.Errorf("error generating nonce: %w", err)
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

//DecryptAccessToken decrypts AES-GCM encoded ciphertext, returning a
//byte slice representing an access token
func (e *Encryptor) DecryptAccessToken(ciphertext, keyString string) ([]byte, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext string: %w", err)
	}

	key, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, fmt.Errorf("error decoding key string to bytes: %w", err)
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("error generating gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return nil, fmt.Errorf("error in ciphertext length")
	}

	nonce, ciphertextBytes := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("error opening ciphertext: %w", err)
	}

	return plaintext, nil
}