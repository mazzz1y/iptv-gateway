package listing

import (
	"crypto/sha256"
	"fmt"
)

func GenerateTvgID(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash[:4])
}
