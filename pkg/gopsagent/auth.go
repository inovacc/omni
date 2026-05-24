package gopsagent

import (
	"crypto/hmac"
	"crypto/sha256"
)

// expectedHMAC returns HMAC-SHA256(challenge, key). Both server and client
// compute this — the server emits the challenge, the client returns this
// value, and the server compares with hmac.Equal (constant-time).
func expectedHMAC(key, challenge []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(challenge)
	return mac.Sum(nil)
}
