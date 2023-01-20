package stop

import (
	"context"
)

type ContainerManager interface {
	StopAll(ctx context.Context)
}
