package mcp

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func CanonicalizeAndHash(tool Tool) (string, error) {
	// Use canonical serialization (deterministic field order)
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")

	if err := encoder.Encode(tool); err != nil {
		return "", fmt.Errorf("failed to serialize tool: %w", err)
	}

	hash := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", hash[:]), nil
}
