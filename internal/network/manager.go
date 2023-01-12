package network

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/wmentor/shrun/internal/common"
)

type Manager struct {
	client      *client.Client
	networkName string
}

func NewManager(client *client.Client) (*Manager, error) {
	if client == nil {
		return nil, common.ErrInvalidDockerClient
	}

	mng := &Manager{
		client:      client,
		networkName: fmt.Sprintf("%snetwork", common.GetObjectPrefix()),
	}

	return mng, nil
}

func (mng *Manager) CheckNetworkExists(ctx context.Context) (bool, error) {
	if _, err := mng.fetchNetwork(ctx); err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (mng *Manager) CreateNetwork(ctx context.Context) (string, error) {
	if _, err := mng.fetchNetwork(ctx); err != nil {
		if !errors.Is(err, common.ErrNotFound) {
			return "", err
		}

		opts := types.NetworkCreate{
			CheckDuplicate: true,
			Driver:         "bridge",
		}

		resp, err := mng.client.NetworkCreate(ctx, mng.networkName, opts)
		if err != nil {
			return "", fmt.Errorf("create network %s error: %w", mng.networkName, err)
		}

		return resp.ID, nil
	}

	return "", fmt.Errorf("network %s: %w", mng.networkName, common.ErrAlreadyExists)
}

func (mng *Manager) DeleteNetwork(ctx context.Context) error {
	id, err := mng.fetchNetwork(ctx)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			log.Print("network not found")
		}
		return err
	}

	return mng.client.NetworkRemove(ctx, id)
}

func (mng *Manager) GetNetworkName() string {
	return mng.networkName
}

func (mng *Manager) GetNetworkID(ctx context.Context) (string, error) {
	return mng.fetchNetwork(ctx)
}

func (mng *Manager) fetchNetwork(ctx context.Context) (string, error) {
	filter := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: mng.networkName})
	opts := types.NetworkListOptions{
		Filters: filter,
	}
	list, err := mng.client.NetworkList(ctx, opts)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("network %s: %w", mng.networkName, common.ErrNotFound)
	}
	return list[0].ID, nil
}
