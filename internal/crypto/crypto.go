package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/m4xvel/monetych_bot/internal/logger"
)

type Service struct {
	key []byte
}

func New(keyBase64 string) (*Service, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		logger.Log.Errorw("failed decode string",
			"err", err)
		return nil, err
	}

	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes (AES-256)")
	}

	return &Service{key: key}, nil
}

func (s *Service) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		logger.Log.Errorw("failed to creates and returns a new cipher.Block",
			"err", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Log.Errorw("failed to returns block cipher",
			"err", err)
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Log.Errorw("failed to returns the number of bytes copied",
			"err", err)
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return append(nonce, ciphertext...), nil
}

func (s *Service) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		logger.Log.Errorw("failed to creates and returns a new cipher.Block",
			"err", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Log.Errorw("failed to returns block cipher",
			"err", err)
		return nil, err
	}

	ns := gcm.NonceSize()
	if len(ciphertext) < ns {
		return nil, errors.New("invalid ciphertext")
	}

	nonce := ciphertext[:ns]
	data := ciphertext[ns:]

	return gcm.Open(nil, nonce, data, nil)
}
