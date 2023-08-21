package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	in "github.com/wmentor/shrun/internal/cases/init"
	"github.com/wmentor/shrun/internal/common"
	"github.com/wmentor/shrun/internal/entities"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandPull)(nil)
)

type CommandInit struct {
	command  *cobra.Command
	cli      *client.Client
	settings entities.ExportFileSettings
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

	cc.Flags().BoolVar(&ci.settings.NoGoProxy, "disable-go-proxy", false, "disable go proxy (default false)")
	cc.Flags().BoolVar(&common.EnableSSL, "ssl", false, "enable ssl")
	cc.Flags().BoolVar(&common.EnableStrictHBA, "strict-hba", false, "enable strictUserHBA")
	cc.Flags().IntVar(&ci.settings.Repfactor, "repfactor", 1, "replication factor (default 1)")
	cc.Flags().StringVar(&ci.settings.Topology, "topology", "cross", "cluster topology (cross or manual, cross as default)")
	cc.Flags().IntVar(&ci.settings.EtcdCount, "etcd-count", 1, "etcd instance count (default 1)")
	cc.Flags().StringVar(&ci.settings.LogLevel, "log-level", "debug", "log level (default debug)")
	cc.Flags().BoolVar(&ci.settings.Debug, "debug", false, "enable debug mode")

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

	myCase, err := in.NewCase(imageManager, ci.settings)
	if err != nil {
		return err
	}

	return myCase.Exec(cc.Context())
}
