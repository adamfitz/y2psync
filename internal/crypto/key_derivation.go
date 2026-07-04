package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	SaltLength    = 16
	Argon2Time    = 3
	Argon2Memory  = 64 * 1024
	Argon2Threads = 4
	KeyLength     = 32
)

func DeriveSyncGroupKey(masterKey string, salt []byte) []byte {
	return argon2.IDKey([]byte(masterKey), salt, Argon2Time, Argon2Memory, Argon2Threads, KeyLength)
}

func DeriveRendezvousTag(masterKey string) string {
	h := sha256.Sum256([]byte("rendezvous:" + masterKey))
	return hex.EncodeToString(h[:])
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}
	return salt, nil
}
