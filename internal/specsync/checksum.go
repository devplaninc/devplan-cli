package specsync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// calculateChecksum computes the SHA-256 checksum of a file
func calculateChecksum(filePath string) (string, []byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return CalculateChecksumBytes(data), data, nil
}

// CalculateChecksumBytes computes the SHA-256 checksum of byte data
func CalculateChecksumBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
