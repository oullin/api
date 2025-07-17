package auth

import "strings"

func SafeDisplay(secret string) string {
	var prefixLen int
	visibleChars := 10

	if strings.HasPrefix(secret, PublicKeyPrefix) {
		prefixLen = len(PublicKeyPrefix)
	} else {
		prefixLen = len(SecretKeyPrefix)
	}

	if len(secret) <= prefixLen+visibleChars {
		return secret
	}

	return secret[:prefixLen+visibleChars] + "..."
}
