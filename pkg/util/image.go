package util

import (
	"strings"
)

// ExtractRegistryAndTag extracts registry and tag from image name
func ExtractRegistryAndTag(imageName string) (string, string) {
	parts := strings.Split(imageName, "/")
	registry := "unknown"
	tag := "latest"

	if len(parts) >= 2 {
		registry = parts[0]
		// Extract tag from the last part
		lastPart := parts[len(parts)-1]
		if strings.Contains(lastPart, ":") {
			tagParts := strings.Split(lastPart, ":")
			if len(tagParts) >= 2 {
				tag = tagParts[1]
			}
		}
	}

	return registry, tag
}
