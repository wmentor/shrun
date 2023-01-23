package main

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd/shrun/cmd"
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
	baseCommand.AddCommand(cmd.NewCommandInit(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandNodes(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandPull(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandShell(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStart(cli).Command())
	baseCommand.AddCommand(cmd.NewCommandStop(cli).Command())

	log.Printf("platform: %s/%s", runtime.GOOS, runtime.GOARCH)

	baseCommand.ExecuteContext(context.Background())
}
