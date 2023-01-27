package doc

import (
	"context"
)

type Builder interface {
	BuildImage(ctx context.Context, dockerfile string, tag string) error
}
