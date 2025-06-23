package registry

import (
	"context"
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/rnathani/kubectl-analyze-images/pkg/types"
)

// Client represents an OCI registry client
type Client struct {
	// For now, we'll implement basic functionality
	// In later phases, we'll add proper OCI client implementation
}

// NewClient creates a new registry client
func NewClient() *Client {
	return &Client{}
}

// GetImageInfo fetches image information from registry
func (c *Client) GetImageInfo(ctx context.Context, imageName string) (*types.Image, error) {
	// Parse image reference
	ref, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}

	// Extract registry and tag
	registry := reference.Domain(ref)
	tag := "latest"
	if tagged, ok := ref.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	// For Phase 1, we'll return mock data
	// In later phases, we'll implement actual registry queries
	image := &types.Image{
		Name:     imageName,
		Size:     1024 * 1024 * 100, // Mock 100MB size
		Registry: registry,
		Tag:      tag,
		Layers: []types.Layer{
			{
				Digest: "sha256:mock-layer-1",
				Size:   1024 * 1024 * 50, // Mock 50MB layer
			},
			{
				Digest: "sha256:mock-layer-2",
				Size:   1024 * 1024 * 50, // Mock 50MB layer
			},
		},
	}

	return image, nil
}

// GetImageSize returns the total size of an image
func (c *Client) GetImageSize(ctx context.Context, imageName string) (int64, error) {
	image, err := c.GetImageInfo(ctx, imageName)
	if err != nil {
		return 0, err
	}
	return image.Size, nil
}

// IsImageAccessible checks if an image is accessible
func (c *Client) IsImageAccessible(ctx context.Context, imageName string) bool {
	// For Phase 1, assume all images are accessible
	// In later phases, we'll implement actual checks
	return true
}
