package doc

import (
	"context"

	"github.com/wmentor/shrun/internal/entities"
)

type ContainerManager interface {
	CreateAndStart(ctx context.Context, css entities.ContainerStartSettings) (string, error)
	Exec(ctx context.Context, conID string, command string, username string) (int, error)
	RemoveContainer(ctx context.Context, name string) error
}
