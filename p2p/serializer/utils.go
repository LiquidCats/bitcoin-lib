package serializer

import (
	"crypto/sha256"
	"fmt"
)

func Checksum(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])

	return second[:]
}

// maxNetAddressPayload returns the max payload size for a bitcoin NetAddress
// based on the protocol version.
func MaxNetAddressPayload(pver uint32) uint32 {
	// Services 8 bytes + ip 16 bytes + port 2 bytes.
	plen := uint32(26)

	// NetAddressTimeVersion added a timestamp field.
	if pver >= NetAddressTimeVersion {
		// Timestamp 4 bytes.
		plen += 4
	}

	return plen
}

// validateUserAgent checks userAgent length against MaxUserAgentLen
func ValidateUserAgent(f string, userAgent string) error {
	if len(userAgent) > MaxUserAgentLen {
		return fmt.Errorf("%s: user agent too long [len %v, max %v]", f, len(userAgent), MaxUserAgentLen)
	}
	return nil
}
