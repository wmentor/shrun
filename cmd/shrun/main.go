package main

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd/shrun/cmd"
	"github.com/wmentor/shrun/internal/common"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithTimeout(time.Minute*3))
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	baseCommand := &cobra.Command{
		Use:     "shrun",
		Short:   "manage shardman cluster for dev",
		Version: common.Version,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if common.WorkArch != common.ArchAmd64 && common.WorkArch != common.ArchArm64 {
				return common.ErrInvalidArch
			}
			if runtime.GOARCH != common.WorkArch {
				log.Printf("build platform: linux/%s", common.WorkArch)
			}
			return nil
		},
	}

	baseCommand.CompletionOptions.HiddenDefaultCmd = true

	baseCommand.AddCommand(cmd.NewCommandBuild(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandCore(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandDoc(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandGoBuilder(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandGoTpc(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandInit(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandNodes(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandPSQL(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandPause(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandPull(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandResume(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandShell(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStart(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStop(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandClean(cli).Command())

	log.Printf("host platform: %s/%s", runtime.GOOS, runtime.GOARCH)

	baseCommand.PersistentFlags().StringVar(&common.ObjectPrefix, "prefix", "shr", "instance prefix")
	baseCommand.PersistentFlags().StringVar(&common.WorkArch, "arch", common.GetDefaultArch(), "build arch (amd64 or arm64)")
	baseCommand.PersistentFlags().StringVar(&common.ClusterName, "cluster", "cluster0", "Shardman cluster name")
	baseCommand.PersistentFlags().IntVar(&common.PgVersion, "pg-major", 14, "PostgresSQL major version (default 14)")

	baseCommand.ExecuteContext(context.Background())
}
