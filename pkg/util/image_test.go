package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRegistryAndTag(t *testing.T) {
	tests := []struct {
		name             string
		imageName        string
		expectedRegistry string
		expectedTag      string
	}{
		{
			name:             "docker.io with tag",
			imageName:        "docker.io/library/nginx:1.21",
			expectedRegistry: "docker.io",
			expectedTag:      "1.21",
		},
		{
			name:             "gcr.io with version tag",
			imageName:        "gcr.io/project/image:v1.0",
			expectedRegistry: "gcr.io",
			expectedTag:      "v1.0",
		},
		{
			name:             "quay.io with latest tag",
			imageName:        "quay.io/organization/app:latest",
			expectedRegistry: "quay.io",
			expectedTag:      "latest",
		},
		{
			name:             "docker.io without tag",
			imageName:        "docker.io/library/nginx",
			expectedRegistry: "docker.io",
			expectedTag:      "latest",
		},
		{
			name:             "single component no registry",
			imageName:        "nginx",
			expectedRegistry: "docker.io",
			expectedTag:      "latest",
		},
		{
			name:             "single component with tag",
			imageName:        "nginx:1.21",
			expectedRegistry: "docker.io",
			expectedTag:      "1.21",
		},
		{
			name:             "empty string",
			imageName:        "",
			expectedRegistry: "docker.io",
			expectedTag:      "latest",
		},
		{
			name:             "custom registry with port",
			imageName:        "registry.company.com/team/app:v2",
			expectedRegistry: "registry.company.com",
			expectedTag:      "v2",
		},
		{
			name:             "nested path",
			imageName:        "gcr.io/proj/team/service:1.0",
			expectedRegistry: "gcr.io",
			expectedTag:      "1.0",
		},
		{
			name:             "special characters in tag",
			imageName:        "docker.io/app:v1.0-alpha",
			expectedRegistry: "docker.io",
			expectedTag:      "v1.0-alpha",
		},
		{
			name:             "localhost with port",
			imageName:        "localhost:5000/app:dev",
			expectedRegistry: "localhost:5000",
			expectedTag:      "dev",
		},
		{
			name:             "digest instead of tag",
			imageName:        "docker.io/nginx@sha256:abc123",
			expectedRegistry: "docker.io",
			expectedTag:      "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, tag := ExtractRegistryAndTag(tt.imageName)
			assert.Equal(t, tt.expectedRegistry, registry)
			assert.Equal(t, tt.expectedTag, tag)
		})
	}
}
