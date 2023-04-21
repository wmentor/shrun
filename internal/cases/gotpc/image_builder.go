package gotpc

import (
	"context"
)

type ImageBuilder interface {
	BuildImage(ctx context.Context, dockerfile string, tag string) error
}
