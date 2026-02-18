package jsonutil

import (
	"net/http"
	"strings"
)

// GetIDFromPath extracts the ID from a RESTful path given a prefix.
// Example: GetIDFromPath(r, "/v1/wallets/") for path "/v1/wallets/user_123" returns "user_123".
func GetIDFromPath(r *http.Request, prefix string) string {
	return strings.TrimPrefix(r.URL.Path, prefix)
}

// GetIDAfter extracts the ID after a specific segment in the path.
// Example: GetIDAfter(r, "intents") for path "/v1/payments/intents/pi_123/confirm" returns "pi_123".
func GetIDAfter(r *http.Request, segment string) string {
	parts := strings.Split(r.URL.Path, "/")
	for i, p := range parts {
		if p == segment && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
