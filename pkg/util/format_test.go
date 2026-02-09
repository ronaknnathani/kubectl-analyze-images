package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "500 bytes",
			bytes:    500,
			expected: "500 B",
		},
		{
			name:     "1023 bytes",
			bytes:    1023,
			expected: "1023 B",
		},
		{
			name:     "1024 bytes",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "1536 bytes",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "1 MB",
			bytes:    1048576,
			expected: "1.0 MB",
		},
		{
			name:     "1 GB",
			bytes:    1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "1 TB",
			bytes:    1099511627776,
			expected: "1.0 TB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytesShort(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0B",
		},
		{
			name:     "500 bytes",
			bytes:    500,
			expected: "500B",
		},
		{
			name:     "1024 bytes",
			bytes:    1024,
			expected: "1K",
		},
		{
			name:     "1 MB",
			bytes:    1048576,
			expected: "1M",
		},
		{
			name:     "1 GB",
			bytes:    1073741824,
			expected: "1G",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytesShort(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
