package clean

import "context"

type Cleaner interface {
	RemoveImage(ctx context.Context, name string, force bool) error
	BuilderPrune(ctx context.Context, all bool) error
}
