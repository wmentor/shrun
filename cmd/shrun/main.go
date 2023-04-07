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

var (
	Version = "0.1"
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
		Version: Version,
	}

	baseCommand.CompletionOptions.HiddenDefaultCmd = true

	baseCommand.AddCommand(cmd.NewCommandBuild(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandDoc(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandGoBuilder(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandInit(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandNodes(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandPSQL(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandPull(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandShell(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStart(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStolon(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStop(cli).Command())

	log.Printf("platform: %s/%s", runtime.GOOS, runtime.GOARCH)

	baseCommand.PersistentFlags().StringVar(&common.ObjectPrefix, "prefix", "shr", "instance prefix")
	baseCommand.PersistentFlags().StringVar(&common.ClusterName, "cluster", "cluster0", "Shardman cluster name")
	baseCommand.PersistentFlags().IntVar(&common.PgVersion, "pg-major", 14, "PostgresSQL major version (default 14)")

	baseCommand.ExecuteContext(context.Background())
}
