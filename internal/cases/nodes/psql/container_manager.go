package psql

import (
	"context"
)

type ContainerManager interface {
	ShellCommand(ctx context.Context, containerName string, username string, workDir string, command []string) error
}
