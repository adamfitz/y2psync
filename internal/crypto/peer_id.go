package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GeneratePeerID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate peer id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
