package image

import (
	"bufio"
	"bytes"
	"context"
	"errors"
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

	"github.com/wmentor/shrun/internal/cases/pull"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/entities"
	"github.com/wmentor/shrun/internal/tmpl"
)

var (
	_ pull.ImageManager = (*Manager)(nil)
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

func (mng *Manager) BuilderPrune(ctx context.Context, all bool) error {
	_, err := mng.client.BuildCachePrune(ctx, types.BuildCachePruneOptions{All: all})
	return err
}

func (mng *Manager) PullImage(ctx context.Context, name string) error {
	opts := types.ImagePullOptions{
		Platform: "linux/" + common.WorkArch,
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
	_, err := mng.getImageId(ctx, name)
	return err
}

func (mng *Manager) BuildImage(ctx context.Context, dockerfile string, tag string) error {
	dir := filepath.Join(common.GetConfigDir(), dockerfile)

	args := make([]string, 0, 10)

	if runtime.GOARCH == common.ArchAmd64 && common.WorkArch == common.ArchAmd64 {
		args = append(args, "build", "--platform", "linux/"+common.WorkArch)
	} else {
		args = append(args, "buildx", "build", "--platform", "linux/"+common.WorkArch)
	}

	args = append(args, "-t", tag, "-f", dir, common.GetDataDir())

	cmd := exec.CommandContext(ctx, "docker", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (mng *Manager) RemoveImage(ctx context.Context, name string, force bool) error {
	imageId, err := mng.getImageId(ctx, name)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return err
	}
	response, err := mng.client.ImageRemove(ctx, imageId, types.ImageRemoveOptions{
		Force:         force,
		PruneChildren: false,
	})
	if err != nil {
		return err
	}
	for _, x := range response {
		log.Println("deleted: ", x.Deleted)
		log.Println("untagged: ", x.Untagged)
	}
	return nil
}

func (mng *Manager) getImageId(ctx context.Context, name string) (string, error) {
	opts := types.ImageListOptions{
		All: true,
	}
	images, err := mng.client.ImageList(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("get image list error: %w", err)
	}
	for _, image := range images {
		if len(image.RepoTags) > 0 && image.RepoTags[0] != "<none>:<none>" {
			if image.RepoTags[0] == name {
				return image.ID, nil
			}
		}
	}
	return "", common.ErrNotFound
}

type specRecord struct {
	name string
	data []byte
}

func (mng *Manager) ExportFiles(settings entities.ExportFileSettings) error {
	files := []specRecord{
		{common.SpecFile, tmpl.SrcSpec},
		{common.RcLocalFile, tmpl.SrcRcLocal},
		{common.DockerfileGoBuilder, tmpl.SrcGoBuilder},
		{common.DockerfileGoTpc, tmpl.SrcGoTpc},
		{common.DockerfilePgBuildEnv, tmpl.SrcPgBuildEnv},
		{common.DockerfilePgDestEnv, tmpl.SrcPgDestEnv},
		{common.DockerfileSdmNode, tmpl.SrcSdmNode},
		{common.DockerfileShardman, tmpl.SrcShardman},
		{common.DockerfilePgDoc, tmpl.SrcPgDoc},
		{common.DockerfileStolonInt, tmpl.SrcStolonInt},
		{common.GetObjectPrefix() + ".env", tmpl.EnvFile},
	}

	for _, rec := range files {
		data := string(rec.data)

		if settings.NoGoProxy {
			data = strings.ReplaceAll(data, "ENV GOPROXY", "#ENV GOPROXY")
			data = strings.ReplaceAll(data, "ENV GONOPROXY", "#ENV GONOPROXY")
		}

		if common.WorkArch == common.ArchArm64 {
			data = strings.ReplaceAll(data, "{{ UbuntuImage }}", "arm64v8/ubuntu:20.04")
			data = strings.ReplaceAll(data, "{{ EtcdImage }}", "quay.io/coreos/etcd:v3.5.8-arm64")
		} else {
			data = strings.ReplaceAll(data, "{{ UbuntuImage }}", "ubuntu:20.04")
			data = strings.ReplaceAll(data, "{{ EtcdImage }}", "quay.io/coreos/etcd:v3.5.8")
		}

		data = strings.ReplaceAll(data, "{{ Repfactor }}", strconv.Itoa(settings.Repfactor))
		data = strings.ReplaceAll(data, "{{ PlacementPolicy }}", settings.Topology)
		data = strings.ReplaceAll(data, "{{ ClusterName }}", common.ClusterName)
		data = strings.ReplaceAll(data, "{{ LogLevel }}", settings.LogLevel)
		data = strings.ReplaceAll(data, "{{ PgMajor }}", strconv.Itoa(common.PgVersion))
		data = strings.ReplaceAll(data, "{{ SdmNodeImage }}", common.GetSdmNodeImageName())

		data = strings.ReplaceAll(data, "{{ Arch }}", common.WorkArch)

		data = mng.handleDebug(data, settings.Debug)

		maker := bytes.NewBuffer(nil)
		for i := 1; i <= settings.EtcdCount; i++ {
			if i > 1 {
				maker.WriteRune(',')
			}
			fmt.Fprintf(maker, "http://%se%d:2379", common.GetObjectPrefix(), i)
		}
		data = strings.ReplaceAll(data, "{{ EtcdList }}", maker.String())

		if err := mng.exportFile(filepath.Join(common.GetConfigDir(), rec.name), []byte(data)); err != nil {
			return fmt.Errorf("export %s error: %w", rec.name, err)
		}
	}

	return nil
}

func (mng *Manager) handleDebug(data string, debug bool) string {
	if debug {
		data = strings.ReplaceAll(data, "{{ CopyDebugTool }}", common.CopyDebugToolCmd)
		data = strings.ReplaceAll(data, "{{ Build }}", common.BuildDebug)

		return data
	}
	data = strings.ReplaceAll(data, "{{ CopyDebugTool }}", "")
	data = strings.ReplaceAll(data, "{{ Build }}", common.BuildDefault)

	return data
}

func (mng *Manager) exportFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}
