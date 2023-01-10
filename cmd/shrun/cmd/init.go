package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandPull)(nil)
)

type CommandInit struct {
	command     *cobra.Command
	cli         *client.Client
	noGoProxy   bool
	repfactor   int
	topology    string
	clusterName string
	etcdCount   int
	logLevel    string
	pgMajor     int
}

func NewCommandInit(cli *client.Client) *CommandInit {
	ci := &CommandInit{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "init",
		Short: "init config and spec file template",
		RunE:  ci.exec,
	}

	cc.Flags().BoolVar(&ci.noGoProxy, "disable-go-proxy", false, "disable go proxy (default false)")
	cc.Flags().IntVar(&ci.repfactor, "repfactor", 1, "replication factor (default 1)")
	cc.Flags().StringVar(&ci.topology, "topology", "cross", "cluster topology (cross or manual, cross as default)")
	cc.Flags().IntVar(&ci.etcdCount, "etcd-count", 1, "etcd instance count (default 1)")
	cc.Flags().IntVar(&ci.pgMajor, "pg-major", 14, "postgres major version (default 14)")
	cc.Flags().StringVar(&ci.logLevel, "log-level", "debug", "log level (default debug)")
	cc.Flags().StringVar(&ci.clusterName, "cluster", "cluster0", "cluster name (default cluster0)")

	ci.command = cc

	return ci
}

func (ci *CommandInit) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandInit) exec(cc *cobra.Command, _ []string) error {
	imageManager, err := image.NewManager(ci.cli)
	if err != nil {
		return err
	}

	settings := image.ExportSettings{
		NoGoProxy:   ci.noGoProxy,
		Repfactor:   ci.repfactor,
		Topology:    ci.topology,
		EtcdCount:   ci.etcdCount,
		LogLevel:    ci.logLevel,
		ClusterName: ci.clusterName,
		PgMajor:     ci.pgMajor,
	}

	return imageManager.ExportFiles(settings)
}
