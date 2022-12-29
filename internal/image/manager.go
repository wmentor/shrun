package image

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/image/source"
)

const (
	FileDockerGoBuilder = "Dockerfile.gobuilder"
)

type Manager struct {
	client *client.Client
}

func NewManager(client *client.Client) (*Manager, error) {
	if client == nil {
		return nil, common.ErrInvalidDockerClient
	}

	mng := &Manager{
		client: client,
	}

	return mng, nil
}

func (mng *Manager) PullImage(ctx context.Context, name string) error {
	opts := types.ImagePullOptions{
		Platform: "linux/amd64",
	}

	src := fmt.Sprintf("docker.io/library/%s", name)

	reader, err := mng.client.ImagePull(ctx, src, opts)
	if err != nil {
		return fmt.Errorf("pull %s error: %w", name, err)
	}

	br := bufio.NewReader(reader)

	for {
		str, err := br.ReadString('\n')
		if err != nil && str == "" {
			if err != io.EOF {
				return err
			}
			break
		}
		log.Print(str)
	}

	return nil
}

func (mng *Manager) CheckImageExists(ctx context.Context, name string) error {
	opts := types.ImageListOptions{
		All: true,
	}
	images, err := mng.client.ImageList(ctx, opts)
	if err != nil {
		return fmt.Errorf("get image list error: %w", err)
	}

	for _, image := range images {
		if len(image.RepoTags) > 0 && image.RepoTags[0] != "<none>:<none>" {
			if image.RepoTags[0] == name {
				return nil
			}
		}
	}

	return common.ErrNotFound
}

type specRecord struct {
	name string
	data []byte
}

func (mng *Manager) ExportFiles() error {
	files := []specRecord{
		{FileDockerGoBuilder, source.SrcGoBuilder},
	}

	for _, rec := range files {
		if err := mng.exportFile(filepath.Join(common.GetConfigDir(), rec.name), rec.data); err != nil {
			return fmt.Errorf("export %s error: %w", rec.name, err)
		}
	}

	return nil
}

func (mng *Manager) exportFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}
