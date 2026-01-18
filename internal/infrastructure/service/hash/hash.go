package hashservice

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type Config struct {
	saltLength int `env:"PASSWORD_SALT_LENGTH"`
}
type HasherService struct {
	cfg *Config
}

func New(cfg *Config) *HasherService {
	return &HasherService{
		cfg: cfg,
	}
}

// Hash generates a hash using sha256 and return base64 string
func (h *HasherService) Hash(str string) (hash string, err error) {
	const op = "hash.Hash()"

	hasher := sha256.New()
	if _, err = hasher.Write([]byte(str)); err != nil {
		return "", apperrors.Wrap(op, err)
	}
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

// Salt generate random salt
func (h *HasherService) Salt() ([]byte, error) {
	const op = "hash.Salt()"

	saltLen := h.cfg.saltLength
	if saltLen == 0 {
		saltLen = 32
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, apperrors.Wrap(op, err)
	}
	return salt, nil
}

// HashWithSalt generates a hash with salt using sha256 and return base64 string
func (h *HasherService) HashWithSalt(str string, salt []byte) (hash string, err error) {
	const op = "hash.HashWithSalt()"

	hasher := sha256.New()
	data := append([]byte(str), salt...)
	if _, err = hasher.Write(data); err != nil {
		return "", apperrors.Wrap(op, err)
	}

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

// ComparePasswords -
func (h *HasherService) ComparePasswords(storedHash, password string, salt string) (bool, error) {
	const op = "hash.ComparePasswords()"

	byteSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return false, apperrors.Wrap(op, err)
	}

	computedHash, err := h.HashWithSalt(password, byteSalt)
	if err != nil {
		return false, apperrors.Wrap(op, err)
	}

	return storedHash == computedHash, nil
}
