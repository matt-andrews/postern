package harness

import (
	"crypto/rand"
	"encoding/hex"
)

// generate a unique name for testing purposes
func UniqueName(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}
