package clean

import "context"

type ImageRemover interface {
	RemoveImage(ctx context.Context, name string, force bool) error
}
