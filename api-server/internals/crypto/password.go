package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLen      uint32
}

type ParsedPassword struct {
	PasswordHash []byte
	Salt         []byte
	Params
}

func generateSalt(size uint32) ([]byte, error) {
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func HashPassword(password string, salt []byte, params Params) []byte {
	key := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLen)
	return key
}

func HashAndEncodePassword(password string) (string, error) {
	salt, err := generateSalt(16)
	if err != nil {
		return "", err
	}
	params := Params{
		Memory:      20 * 1024,
		Iterations:  2,
		Parallelism: 1,
		KeyLen:      32,
	}
	key := HashPassword(password, salt, params)

	b64Key := base64.RawStdEncoding.EncodeToString(key)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)

	hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Key)
	return hash, nil
}


var ErrInvalidHash = errors.New("invalid hash")

func ParsePassword(password string) (ParsedPassword, error) {
	parts := strings.Split(password, "$")

	if len(parts) != 6 {
		return ParsedPassword{}, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return ParsedPassword{}, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}

	if version != argon2.Version {
		return ParsedPassword{}, ErrInvalidHash
	}

	if len(parts) != 6 {
		return ParsedPassword{}, ErrInvalidHash
	}
	config := strings.Split(parts[3], ",")
	if len(config) != 3 {
		return ParsedPassword{}, ErrInvalidHash
	}
	var (
		memory      uint32
		iterations  uint32
		parallelism uint8
	)
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}
	passwordHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return ParsedPassword{}, ErrInvalidHash
	}
	return ParsedPassword{
		PasswordHash: passwordHash,
		Salt:         salt,
		Params: Params{
			Memory:      memory,
			Iterations:  iterations,
			Parallelism: parallelism,
			KeyLen:      uint32(len(passwordHash)),
		},
	}, nil
}

// ""
// argon2id
// v=""
// m=,t=,p=
// ""
// ""
