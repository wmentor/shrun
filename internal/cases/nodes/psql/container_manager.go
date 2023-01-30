package psql

import (
	"context"
)

type ContainerManager interface {
	ShellCommand(ctx context.Context, containerName string, username string, command []string) error
}
