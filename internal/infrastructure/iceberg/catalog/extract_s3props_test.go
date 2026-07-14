/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package catalog

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtractS3Props verifies that extractS3Props correctly collects all
// "s3.*" DSN query parameters and ignores unrelated parameters.
func TestExtractS3Props(t *testing.T) {
	tests := []struct {
		name     string
		input    url.Values
		expected map[string]string
	}{
		{
			name: "all standard s3 params",
			input: url.Values{
				"s3.endpoint":                 {"http://minio:9000"},
				"s3.access-key-id":            {"minioadmin"},
				"s3.secret-access-key":        {"minioadmin"},
				"s3.region":                   {"us-east-1"},
				"s3.force-virtual-addressing": {"false"},
			},
			expected: map[string]string{
				"s3.endpoint":                 "http://minio:9000",
				"s3.access-key-id":            "minioadmin",
				"s3.secret-access-key":        "minioadmin",
				"s3.region":                   "us-east-1",
				"s3.force-virtual-addressing": "false",
			},
		},
		{
			name: "s3 params mixed with non-s3 params",
			input: url.Values{
				"token":                {"mytoken"},
				"prefix":               {"myprefix"},
				"s3.endpoint":          {"http://localhost:9000"},
				"s3.access-key-id":     {"ak"},
				"s3.secret-access-key": {"sk"},
			},
			expected: map[string]string{
				"s3.endpoint":          "http://localhost:9000",
				"s3.access-key-id":     "ak",
				"s3.secret-access-key": "sk",
			},
		},
		{
			name:     "no s3 params returns nil",
			input:    url.Values{"token": {"tok"}, "prefix": {"p"}},
			expected: nil,
		},
		{
			name:     "empty values returns nil",
			input:    url.Values{},
			expected: nil,
		},
		{
			name: "s3 with session token",
			input: url.Values{
				"s3.access-key-id":     {"ak"},
				"s3.secret-access-key": {"sk"},
				"s3.session-token":     {"st"},
			},
			expected: map[string]string{
				"s3.access-key-id":     "ak",
				"s3.secret-access-key": "sk",
				"s3.session-token":     "st",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractS3Props(tt.input)
			if tt.expected == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.expected, map[string]string(got))
			}
		})
	}
}
