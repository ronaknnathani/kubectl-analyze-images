package util

import (
	"strings"
)

// ExtractRegistryAndTag extracts registry and tag from image name
func ExtractRegistryAndTag(imageName string) (registry, tag string) {
	parts := strings.Split(imageName, "/")
	registry = "docker.io"
	tag = "latest"

	// Get the last part which contains the image name and possibly tag
	lastPart := parts[len(parts)-1]

	// Extract tag from the last part if present
	if strings.Contains(lastPart, ":") {
		tagParts := strings.Split(lastPart, ":")
		if len(tagParts) >= 2 {
			tag = tagParts[1]
		}
	}

	// Determine registry based on number of parts
	if len(parts) >= 2 {
		registry = parts[0]
	}

	return registry, tag
}
