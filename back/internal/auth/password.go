package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime      uint32 = 2 //сколько проходов по памяти 
	argonMemory    uint32 = 19 * 1024 
	argonThreads   uint8  = 1
	argonKeyLength uint32 = 32
	saltLength            = 16
)

func hashPassword(password string) (string, error) {
	salt:= make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil{
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLength,
	)
	return fmt.Sprintf(
		"argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawURLEncoding.EncodeToString(salt),
		base64.RawURLEncoding.EncodeToString(hash),
	), nil

}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")

	if len(parts) != 5 ||
		parts[0] != "argon2id" ||
		parts[1] != "v=19" ||
		parts[2] != "m=19456,t=2,p=1" {
		return false
	}

	encoding := base64.RawURLEncoding

	salt, err := encoding.DecodeString(parts[3])
	if err != nil || len(salt) != saltLength {
		return false
	}

	expectedHash, err := encoding.DecodeString(parts[4])
	if err != nil || len(expectedHash) != int(argonKeyLength) {
		return false
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLength,
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}


