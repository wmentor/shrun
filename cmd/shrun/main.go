package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	netOpts := types.NetworkListOptions{}

	nets, err := cli.NetworkList(ctx, netOpts)
	if err != nil {
		panic(err)
	}

	fmt.Println("Networks:")
	for _, cnet := range nets {
		fmt.Printf("%s %s %s\n", cnet.ID, cnet.Name, cnet.Driver)
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
}
