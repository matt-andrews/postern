package harness

import (
	"crypto/rand"
	"encoding/hex"
)

// UniqueName returns prefix plus a random suffix, e.g. "orders-3fa9c1d2".
// Giving every test its own queue or topic keeps tests from seeing
// each other's messages.
func UniqueName(prefix string) string {
	b := make([]byte, 4)
	rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}
