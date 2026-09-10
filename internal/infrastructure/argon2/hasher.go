package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHashFormat   = errors.New("invalid argon2 hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

type Argon2Hasher struct {
	saltLength uint32
	keyLength  uint32
	time       uint32
	memory     uint32
	threads    uint8
}

func New(saltLength, keyLength, time, memory uint32, threads uint8) Argon2Hasher {
	return Argon2Hasher{
		saltLength: saltLength,
		keyLength:  keyLength,
		time:       time,
		memory:     memory,
		threads:    threads,
	}
}

func NewDefault() Argon2Hasher {
	return New(16, 32, 1, 64*1024, 4)
}

func (a Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, a.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		a.time,
		a.memory,
		a.threads,
		a.keyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		a.memory,
		a.time,
		a.threads,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

func (a Argon2Hasher) Verify(password, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	candidate := argon2.IDKey(
		[]byte(password),
		salt,
		params.time,
		params.memory,
		params.threads,
		uint32(len(hash)),
	)

	if subtle.ConstantTimeCompare(candidate, hash) == 1 {
		return true, nil
	}
	return false, nil
}

type params struct {
	memory  uint32
	time    uint32
	threads uint8
}

func decodeHash(encoded string) (params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return params{}, nil, nil, ErrInvalidHashFormat
	}

	if parts[1] != "argon2id" {
		return params{}, nil, nil, ErrInvalidHashFormat
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return params{}, nil, nil, ErrInvalidHashFormat
	}
	if version != argon2.Version {
		return params{}, nil, nil, ErrIncompatibleVersion
	}

	var p params
	if _, err := fmt.Sscanf(
		parts[3], "m=%d,t=%d,p=%d",
		&p.memory, &p.time, &p.threads,
	); err != nil {
		return params{}, nil, nil, ErrInvalidHashFormat
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params{}, nil, nil, ErrInvalidHashFormat
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params{}, nil, nil, ErrInvalidHashFormat
	}

	return p, salt, hash, nil
}
