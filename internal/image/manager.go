package image

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/image/source"
)

const (
	SpecFile             = "sdmspec.json"
	RcLocal              = "rc.local"
	DockerfileGoBuilder  = "Dockerfile.gobuilder"
	DockerfilePgBuildEnv = "Dockerfile.pgbuildenv"
	DockerfilePgDestEnv  = "Dockerfile.pgdestenv"
	DockerfilePgDoc      = "Dockerfile.pgdoc"
	DockerfileSdmNode    = "Dockerfile.sdmnode"
	DockerfileShardman   = "Dockerfile.shardman"
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
	if strings.Contains(name, "/") {
		src = fmt.Sprintf("docker.io/%s", name)
	}

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

func (mng *Manager) BuildImage(ctx context.Context, dockerfile string, tag string) error {
	dir := filepath.Join(common.GetConfigDir(), dockerfile)

	args := make([]string, 0, 10)

	if runtime.GOARCH == "amd64" {
		args = append(args, "build", "--platform", "linux/amd64")
	} else {
		args = append(args, "buildx", "build", "--platform", "linux/amd64")
	}

	args = append(args, "-t", tag, "-f", dir, common.GetDataDir())

	cmd := exec.CommandContext(ctx, "docker", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type specRecord struct {
	name string
	data []byte
}

type ExportSettings struct {
	NoGoProxy   bool
	Repfactor   int
	Topology    string
	LogLevel    string
	ClusterName string
	EtcdCount   int
	PgMajor     int
}

func (mng *Manager) ExportFiles(settings ExportSettings) error {
	files := []specRecord{
		{SpecFile, source.SrcSpec},
		{RcLocal, source.SrcRcLocal},
		{DockerfileGoBuilder, source.SrcGoBuilder},
		{DockerfilePgBuildEnv, source.SrcPgBuildEnv},
		{DockerfilePgDestEnv, source.SrcPgDestEnv},
		{DockerfileSdmNode, source.SrcSdmNode},
		{DockerfileShardman, source.SrcShardman},
		{DockerfilePgDoc, source.SrcPgDoc},
	}

	for _, rec := range files {
		data := string(rec.data)

		if settings.NoGoProxy {
			data = strings.ReplaceAll(data, "ENV GOPROXY", "#ENV GOPROXY")
			data = strings.ReplaceAll(data, "ENV GONOPROXY", "#ENV GONOPROXY")
		}

		data = strings.ReplaceAll(data, "{{ Repfactor }}", strconv.Itoa(settings.Repfactor))
		data = strings.ReplaceAll(data, "{{ PlacementPolicy }}", settings.Topology)
		data = strings.ReplaceAll(data, "{{ ClusterName }}", settings.ClusterName)
		data = strings.ReplaceAll(data, "{{ LogLevel }}", settings.LogLevel)
		data = strings.ReplaceAll(data, "{{ PgMajor }}", strconv.Itoa(settings.PgMajor))

		maker := bytes.NewBuffer(nil)
		for i := 1; i <= settings.EtcdCount; i++ {
			if i > 1 {
				maker.WriteRune(',')
			}
			fmt.Fprintf(maker, "http://shr_etcd_%d:2379", i)
		}
		data = strings.ReplaceAll(data, "{{ EtcdList }}", maker.String())

		if err := mng.exportFile(filepath.Join(common.GetConfigDir(), rec.name), []byte(data)); err != nil {
			return fmt.Errorf("export %s error: %w", rec.name, err)
		}
	}

	return nil
}

func (mng *Manager) exportFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}
