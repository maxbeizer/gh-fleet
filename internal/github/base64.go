package github

import (
	"encoding/base64"
	"strings"
)

func decodeBase64(s string) (string, error) {
	// GitHub API returns base64 with newlines
	cleaned := strings.ReplaceAll(s, "\n", "")
	b, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
