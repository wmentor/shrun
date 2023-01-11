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

	log.Printf("platform: %s/%s", runtime.GOOS, runtime.GOARCH)

	baseCommand.ExecuteContext(context.Background())

	/*

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srvNetwork, err := network.NewManager(cli)
		if err != nil {
			log.Print(err)
			return
		}

		found, err := srvNetwork.CheckNetworkExists(ctx)
		if err != nil {
			log.Printf("check network exists error: %v", err)
			return
		}

		if found {
			log.Println("network found")
		} else {
			log.Println("network not found. create it")
			if _, err := srvNetwork.CreateNetwork(ctx); err != nil {
				log.Printf("create network exists error: %v", err)
				return
			}
		}

		found, err = srvNetwork.CheckNetworkExists(ctx)
		if err != nil || !found {
			log.Printf("check network exists error: %v", err)
			return
		}

		if err = srvNetwork.DeleteNetwork(ctx); err != nil {
			log.Printf("delete network error: %v", err)
			return
		}

		netOpts := types.NetworkListOptions{}

		nets, err := cli.NetworkList(ctx, netOpts)
		if err != nil {
			panic(err)
		}

		fmt.Println("Networks:")
		for _, cnet := range nets {
			fmt.Printf("%s %s %s %t\n", cnet.ID, cnet.Name, cnet.Driver, cnet.Internal)
		}

		fmt.Println("Images:")

		imgOpt := types.ImageListOptions{All: true}

		images, err := cli.ImageList(ctx, imgOpt)
		if err != nil {
			panic(err)
		}

		for _, image := range images {
			if len(image.RepoTags) > 0 && image.RepoTags[0] != "<none>:<none>" {
				fmt.Printf("%s %v %v\n", image.ID, image.Labels, image.RepoTags)
			}
		}

		fmt.Println("Containers:")

		contOpts := types.ContainerListOptions{All: true}

		containers, err := cli.ContainerList(ctx, contOpts)
		if err != nil {
			panic(err)
		}

		for _, container := range containers {
			fmt.Println(container.ID, container.Names, container.State)
		}


		if err = imageManager.ExportFiles(); err != nil {
			panic(err)
		}

		imageManager.BuildImage(ctx, image.DockerfileGoBuilder, "gobuilder:latest")
	*/
}
