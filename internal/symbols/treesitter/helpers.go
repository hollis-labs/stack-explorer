//go:build treesitter

package treesitter

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
)

func hashText(text string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(sum[:])
}

func inferVisibility(name string) string {
	for _, r := range name {
		if unicode.IsLetter(r) {
			if unicode.IsUpper(r) {
				return "public"
			}
			return "private"
		}
	}
	return ""
}
