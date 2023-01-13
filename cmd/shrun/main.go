package main

import (
	"context"
	"log"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd/shrun/cmd"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	baseCommand := &cobra.Command{
		Use:   "shrun",
		Short: "manage shardman cluster for dev",
	}

	baseCommand.AddCommand(cmd.NewCommandPull(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandInit(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandBuild(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStop(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStart(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandShell(cli).Command())

	log.Printf("platform: %s/%s", runtime.GOOS, runtime.GOARCH)

	baseCommand.ExecuteContext(context.Background())
}
