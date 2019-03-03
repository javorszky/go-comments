package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type argon2Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type Argon2 struct {
	params argon2Params
}

func (a *Argon2) Init(p argon2Params) {
	fmt.Println(p)
	a.params = p
}

// GenerateFromPassword takes a plaintext string and generates an encoded has string with params in it
func (a *Argon2) GenerateFromPassword(password string) (encodedHash string, err error) {
	fmt.Println(a.params)
	salt, err := a.GenerateRandomBytes()

	if err != nil {
		return "", err
	}

	fmt.Println(a.params)

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash := argon2.IDKey([]byte(password), salt, a.params.iterations, a.params.memory, a.params.parallelism, a.params.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, a.params.memory, a.params.iterations, a.params.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// GenerateRandomBytes is used to generate salt used by GenerateFromPassword
func (a *Argon2) GenerateRandomBytes() ([]byte, error) {
	b := make([]byte, a.params.saltLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// ComparePasswordAndHash is used to compare a plaintext pw and an encoded pw hash with params inside.
func (a *Argon2) ComparePasswordAndHash(password string, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := a.DecodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// DecodeHash function is used by ComparePasswordAndHash to extract params used by Argon2
func (a *Argon2) DecodeHash(encodedHash string) (p *argon2Params, salt, hash []byte, err error) {
	values := strings.Split(encodedHash, "$")

	if len(values) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(values[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &argon2Params{}
	_, err = fmt.Sscanf(values[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(values[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(values[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
