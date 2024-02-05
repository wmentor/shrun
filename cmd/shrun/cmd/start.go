package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/build"
	"github.com/wmentor/shrun/internal/cases/stop"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/entities"
	"github.com/wmentor/shrun/internal/image"
	"github.com/wmentor/shrun/internal/network"
)

var (
	_ cmd.CobraCommand = (*CommandBuild)(nil)
)

var (
	ErrInvalidNodeCount     = errors.New("invalid nodes count")
	ErrInvalidFreeNodeCount = errors.New("invalid free nodes count")
)

type CommandStart struct {
	command        *cobra.Command
	cli            *client.Client
	nodesCount     int
	freeNodes      int
	skipNodeAdd    bool
	force          bool
	updateDockers  bool
	memoryLimit    string
	withExtensions []string
	cpuLimit       float64
	debug          bool
	mountData      bool
	openShell      bool
	withGrafana    bool
	withData       bool
	withS3         bool
}

func NewCommandStart(cli *client.Client) *CommandStart {
	c := &CommandStart{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "start",
		Short: "start cluster",
		RunE:  c.exec,
	}

	cc.Flags().IntVarP(&c.nodesCount, "nodes", "n", 2, "number of nodes to add to the cluster")
	cc.Flags().BoolVar(&c.skipNodeAdd, "skip-node-add", false, "skip shardmanctl nodes add")
	cc.Flags().BoolVarP(&c.force, "force", "f", false, "force start. (if already started then restart)")
	cc.Flags().BoolVarP(&c.updateDockers, "update", "u", false, "rebuild utils and update dockers")
	cc.Flags().StringVar(&c.memoryLimit, "memory", "", "memory limit")
	cc.Flags().Float64Var(&c.cpuLimit, "cpu", 0, "cpu limit")
	cc.Flags().BoolVar(&c.debug, "debug", false, "enable debug mode")
	cc.Flags().BoolVar(&c.mountData, "mount-data", false, "mount pg data to builddir/pgdata/hostname")
	cc.Flags().BoolVar(&c.openShell, "shell", false, "open shell")
	cc.Flags().BoolVar(&c.withGrafana, "with-grafana", false, "use grafana")
	cc.Flags().BoolVar(&c.withData, "with-schema", false, "generate start data")
	cc.Flags().BoolVar(&c.withS3, "with-s3", false, "start S3 storage (port 9000 and webui 9001)")
	cc.Flags().StringSliceVarP(&c.withExtensions, "with-extension", "e", []string{}, "extensions to create")
	cc.Flags().IntVar(&c.freeNodes, "free-nodes", 0, "number of nodes that should not be added to the cluster")

	c.command = cc

	return c
}

func (c *CommandStart) Command() *cobra.Command {
	return c.command
}

func (c *CommandStart) exec(cc *cobra.Command, _ []string) error {
	if c.nodesCount < 1 {
		return ErrInvalidNodeCount
	}

	if c.freeNodes < 0 {
		return ErrInvalidFreeNodeCount
	}

	ctx := cc.Context()

	networker, err := network.NewManager(c.cli)
	if err != nil {
		return err
	}

	manager, err := container.NewManager(c.cli)
	if err != nil {
		return err
	}

	started, err := networker.CheckNetworkExists(ctx)
	if err != nil {
		return err
	}

	if started {
		if err = c.isStarted(ctx, manager, networker); err != nil {
			return err
		}
	}

	if c.updateDockers {
		if err = c.update(ctx); err != nil {
			return err
		}
	}

	netID, err := networker.CreateNetwork(ctx)
	if err != nil {
		return fmt.Errorf("create network error: %w", err)
	}

	etcdList, err := common.GetEtcdList()
	if err != nil {
		return err
	}

	etcdClusterMaker := strings.Builder{}
	for i := range etcdList {
		hostname := fmt.Sprintf("%se%d", common.GetObjectPrefix(), i+1)
		if i > 0 {
			etcdClusterMaker.WriteRune(',')
		}
		etcdClusterMaker.WriteString(hostname)
		etcdClusterMaker.WriteRune('=')
		etcdClusterMaker.WriteString(fmt.Sprintf("http://%s:2380", hostname))
	}

	containerIDs := map[string]string{}

	for i := range etcdList {
		hostname := fmt.Sprintf("%se%d", common.GetObjectPrefix(), i+1)

		cmdParams := []string{
			"/usr/local/bin/etcd",
			fmt.Sprintf("--name=%s", hostname),
			fmt.Sprintf("--initial-advertise-peer-urls=http://%s:2380", hostname),
			"--listen-peer-urls=http://0.0.0.0:2380",
			fmt.Sprintf("--advertise-client-urls=http://%s:2379", hostname),
			"--listen-client-urls=http://0.0.0.0:2379",
			fmt.Sprintf("--initial-cluster=%s", etcdClusterMaker.String()),
			"--auto-compaction-retention=5m",
		}

		log.Printf("start %s", hostname)

		etcdImage := ""

		if common.WorkArch == common.ArchArm64 {
			etcdImage = fmt.Sprintf("quay.io/coreos/etcd:v%s-%s", common.EtcdVersion, common.ArchArm64)
		} else {
			etcdImage = "quay.io/coreos/etcd:v" + common.EtcdVersion
		}

		opts := entities.ContainerStartSettings{
			Image:     etcdImage,
			Host:      hostname,
			Cmd:       cmdParams,
			NetworkID: netID,
		}

		if id, err := manager.CreateAndStart(ctx, opts); err == nil {
			containerIDs[hostname] = id
		} else {
			log.Println("failed")
		}
	}

	for i := 0; i < c.nodesCount+c.freeNodes; i++ {
		hostname := common.GetNodeName(i + 1)
		log.Printf("start %s", hostname)

		ports := make([]string, 0)
		if c.debug {
			log.Printf("expose port %d --> %d", 40000, 40000+i)
			ports = append(ports, fmt.Sprintf("%d:40000", 40000+i))
		}

		for j := 0; j < common.ExposePortLimit; j++ {
			log.Printf("expose port %d --> %d", common.DefaultPgPort+j, common.DefaultPgPort+i*common.ExposePortLimit+j)
			ports = append(ports, fmt.Sprintf("%d:%d", common.DefaultPgPort+i*common.ExposePortLimit+j, common.DefaultPgPort+j))
		}

		opts := entities.ContainerStartSettings{
			Image:     common.GetNodeContainerName(),
			Host:      hostname,
			NetworkID: netID,
			Ports:     ports,
			Envs:      common.GetEnvs(),
			MountData: c.mountData,
			Debug:     c.debug,
		}

		if c.memoryLimit != "" {
			size, err := units.RAMInBytes(c.memoryLimit)
			if err != nil {
				return fmt.Errorf("invalid memory limit: %v", c.memoryLimit)
			}
			opts.MemoryLimit = size
		}

		if c.cpuLimit != 0 {
			opts.CPU = c.cpuLimit
		}

		if id, err := manager.CreateAndStart(ctx, opts); err == nil {
			containerIDs[hostname] = id
		} else {
			log.Println("failed")
			return err
		}
	}

	node1ID := containerIDs[common.GetNodeName(1)]

	code, err := manager.Exec(ctx, node1ID, "shardmanctl init -f /etc/shardman/sdmspec.json -y", common.PgUser)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("command status code: %d", code)
	}

	if !c.skipNodeAdd {
		maker := bytes.NewBuffer(nil)
		for i := 0; i < c.nodesCount; i++ {
			if i > 0 {
				maker.WriteRune(',')
			}
			fmt.Fprintf(maker, "%sn%d", common.GetObjectPrefix(), i+1)
		}
		code, err = manager.Exec(ctx, node1ID, "shardmanctl nodes add -n "+maker.String(), common.PgUser)
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("command status code: %d", code)
		}

		for _, ext := range c.withExtensions {
			code, err = manager.Exec(ctx, node1ID, fmt.Sprintf("shardmanctl forall --sql 'CREATE EXTENSION IF NOT EXISTS %s'", qi(ext)), common.PgUser)
			if err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("command status code: %d", code)
			}
		}
	}

	log.Printf("mount: %s --> /mntdata", common.GetVolumeDir())

	if c.withS3 {
		if err := c.runS3(ctx, netID, manager); err != nil {
			return err
		}
	}

	common.SaveGrafanaStatus(c.withGrafana)

	if c.withGrafana {
		if err := c.runGrafana(ctx, netID, manager); err != nil {
			return err
		}
	}

	if c.withData && !c.skipNodeAdd {
		log.Print("generate data")
		code, err = manager.Exec(ctx, node1ID, `psql -d "$(shardmanctl getconnstr)" < /var/lib/postgresql/generate.sql`, common.PgUser)
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("command status code: %d", code)
		}
	}

	if c.openShell {
		return manager.ShellCommand(cc.Context(), common.GetNodeName(1), common.PgUser, "", append(common.CmdBash, "-c", "/var/lib/postgresql/.build_info && /bin/bash"))
	}

	return nil
}

func (c *CommandStart) runS3(ctx context.Context, netID string, manager *container.Manager) error {
	hostname := common.GetObjectPrefix() + "s3"
	log.Printf("start %s", hostname)

	ports := []string{"9000:9000", "9001:9001"}

	cmdParams := []string{"server", "/data", "--console-address", ":9001"}

	envs := append(common.GetEnvs(), "MINIO_ROOT_USER=shardman", "MINIO_ROOT_PASSWORD=shardman")

	opts := entities.ContainerStartSettings{
		Image:     "quay.io/minio/minio:latest",
		Host:      hostname,
		NetworkID: netID,
		Ports:     ports,
		Cmd:       cmdParams,
		Envs:      envs,
	}

	_, err := manager.CreateAndStart(ctx, opts)
	return err
}

func (c *CommandStart) isStarted(ctx context.Context, manager *container.Manager, networker *network.Manager) error {
	if c.force {
		sc, err := stop.NewCase(manager, networker)
		if err != nil {
			return err
		}
		if err = sc.Exec(ctx); err != nil {
			return err
		}
	} else {
		return errors.New("already started")
	}

	return nil
}

func (c *CommandStart) update(ctx context.Context) error {
	builder, err := image.NewManager(c.cli)
	if err != nil {
		return fmt.Errorf("image build init error: %w", err)
	}

	myCase, err := build.NewCase(builder)
	if err != nil {
		return fmt.Errorf("rebuild init case error: %w", err)
	}

	return myCase.Exec(ctx)
}

func (c *CommandStart) runGrafana(ctx context.Context, netID string, manager *container.Manager) error {
	hostname := common.GetObjectPrefix() + "prometheus"
	log.Printf("start %s", hostname)

	ports := []string{"9090:9090"}

	opts := entities.ContainerStartSettings{
		Image:     "prometheus:latest",
		Host:      hostname,
		NetworkID: netID,
		Ports:     ports,
		Envs:      common.GetEnvs(),
		MountData: c.mountData,
	}

	if _, err := manager.CreateAndStart(ctx, opts); err != nil {
		return err
	}

	hostname = common.GetObjectPrefix() + "grafana"
	log.Printf("start %s", hostname)

	ports = []string{"3000:3000"}

	envs := common.GetEnvs()

	envs = append(envs, "GF_SECURITY_ADMIN_PASSWORD=shardman", "GF_SECURITY_ADMIN_USER=shardman")

	opts = entities.ContainerStartSettings{
		Image:     "grafana:latest",
		Host:      hostname,
		NetworkID: netID,
		Ports:     ports,
		Envs:      envs,
		MountData: c.mountData,
	}

	if _, err := manager.CreateAndStart(ctx, opts); err != nil {
		return err
	}

	for i := 0; i < c.nodesCount+c.freeNodes; i++ {
		if err := manager.StartPrometheusExporter(ctx, i+1, netID); err != nil {
			return err
		}
	}

	hostname = common.GetObjectPrefix() + "cadvisor"
	log.Printf("start %s", hostname)

	ports = []string{"8080:8080"}

	opts = entities.ContainerStartSettings{
		Image:     "gcr.io/cadvisor/cadvisor:v0.47.2",
		Host:      hostname,
		NetworkID: netID,
		Ports:     ports,
	}

	if _, err := manager.CreateAndStart(ctx, opts); err != nil {
		log.Printf("start cadvisor error: %v", err)
	}

	return nil
}

// PG's quote_identifier. FIXME keywords
func qi(ident string) string {
	var safe = true
	for _, r := range ident {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r == '_')) {
			safe = false
			break
		}
	}
	if safe {
		return ident
	}
	escaper := strings.NewReplacer(`"`, `""`)
	return fmt.Sprintf("\"%s\"", escaper.Replace(ident))
}
