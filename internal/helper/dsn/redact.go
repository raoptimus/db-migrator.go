/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dsn

import (
	"strings"
)

const redactMask = "****"

// sensitiveParams is the canonical list of query-parameter keys whose values
// must be redacted from DSN strings in logs and error messages.
var sensitiveParams = []string{
	"token",
	"credential",
	"s3.secret-access-key",
	"s3.session-token",
	"password",
}

// Redact returns the DSN with sensitive values masked:
//   - userinfo password in scheme://user:PASS@host → scheme://user:****@host
//   - sensitive query-parameter values: token, credential, s3.secret-access-key,
//     s3.session-token, password → ****
//
// It is robust to malformed DSNs: on parse failure it falls back to best-effort
// string-level masking of query parameters without panicking.
func Redact(rawDSN string) string {
	if rawDSN == "" {
		return rawDSN
	}

	// Try structured redaction first.
	redacted, ok := redactStructured(rawDSN)
	if ok {
		return redacted
	}

	// Fallback: best-effort string-level masking of query params.
	return redactParams(rawDSN)
}

// redactStructured attempts structured redaction via string manipulation.
// Returns the redacted string and true on success, or ("", false) if the DSN
// does not contain "://" (i.e. is not a URL-shaped DSN).
func redactStructured(rawDSN string) (string, bool) {
	// Must have a scheme separator.
	schemeEnd := strings.Index(rawDSN, "://")
	if schemeEnd == -1 {
		return "", false
	}

	scheme := rawDSN[:schemeEnd+3] // "scheme://"
	rest := rawDSN[schemeEnd+3:]   // "user:pass@host/db?params"

	// Mask password in userinfo (user:PASS@...).
	atIdx := strings.Index(rest, "@")
	if atIdx != -1 {
		userinfo := rest[:atIdx]
		afterAt := rest[atIdx:] // includes "@"

		colonIdx := strings.Index(userinfo, ":")
		if colonIdx != -1 && colonIdx < len(userinfo)-1 {
			// There is a non-empty password.
			user := userinfo[:colonIdx]
			userinfo = user + ":" + redactMask
		}
		rest = userinfo + afterAt
	}

	result := scheme + rest

	// Mask sensitive query parameters.
	result = redactParams(result)

	return result, true
}

// redactParams masks the values of sensitive query parameters in s using
// string scanning (does not require a fully-valid URL).
func redactParams(s string) string {
	result := s
	for _, key := range sensitiveParams {
		result = maskParam(result, key)
	}

	return result
}

// maskParam replaces the value of URL query parameter key with redactMask.
// Handles both "key=value&next" and "key=value" (end-of-string) forms.
// All occurrences are masked left-to-right.
func maskParam(s, key string) string {
	prefix := key + "="
	var b strings.Builder
	rest := s

	for {
		idx := strings.Index(rest, prefix)
		if idx == -1 {
			b.WriteString(rest)
			break
		}

		// Copy everything up to and including "key=".
		b.WriteString(rest[:idx+len(prefix)])
		b.WriteString(redactMask)

		// Skip the original value up to the next delimiter or end.
		after := rest[idx+len(prefix):]
		end := strings.IndexAny(after, "&# ")
		if end == -1 {
			// Value extends to end of string.
			break
		}

		rest = after[end:]
	}

	return b.String()
}
