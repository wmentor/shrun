package pull

import (
	"context"
)

type ImageManager interface {
	CheckImageExists(ctx context.Context, name string) error
	PullImage(ctx context.Context, name string) error
}
