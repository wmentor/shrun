package gobuilder

import (
	"context"

	"github.com/wmentor/shrun/internal/entities"
)

type ContainerManager interface {
	CreateAndStart(ctx context.Context, css entities.ContainerStartSettings) (string, error)
	ShellCommand(ctx context.Context, containerName string, username string, command []string) error
	RemoveContainer(ctx context.Context, name string) error
}