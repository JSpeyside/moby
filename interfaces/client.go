package interfaces

import (
	"fmt"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/jlgrady1/moby/infrastructure"
	"golang.org/x/net/context"
	"runtime"
	"strings"
)

type MobyClient struct {
	client *client.Client
	logger *infrastructure.Logger
	Quiet  bool
}

func NewMobyClient() (*MobyClient, error) {
	var cli *client.Client
	var err error
	systemOS := runtime.GOOS
	if systemOS == "darwin" {
		// cli, err = client.NewEnvClient()
		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		cli, err = client.NewClient("unix:///var/run/docker.sock", "v1.23", nil, defaultHeaders)
	} else {
		// defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		// cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
		return nil, fmt.Errorf("Unsupported OS `%s`. Only Mac OS X is currently supported", systemOS)
	}
	logger := infrastructure.NewLogger()
	mobyClient := &MobyClient{cli, logger, false}
	return mobyClient, err
}

func (mc MobyClient) CleanImages() error {
	images := mc.listImages()
	for _, i := range images {
		// fmt.Println(i.RepoTags)
		remove := false
		for _, tag := range i.RepoTags {
			tagNames := strings.Split(tag, ":")
			for _, tagName := range tagNames {
				if tagName == "<none>" {
					remove = true
				}
			}
		}
		if remove == true {
			mc.removeImage(i.ID)
		}
	}
	return nil
}

func (mc MobyClient) listImages() []types.Image {
	options := types.ImageListOptions{All: true}
	images, err := mc.client.ImageList(context.Background(), options)
	if err != nil {
		panic(err)
	}
	return images
}

func (mc MobyClient) listContainers() []types.Container {
	options := types.ContainerListOptions{All: true}
	containers, err := mc.client.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}
	return containers
}

func (mc MobyClient) RemoveStoppedContainers() error {
	containers := mc.listContainers()
	stoppedContainers := []types.Container{}
	for _, c := range containers {
		if c.State != "running" {
			stoppedContainers = append(stoppedContainers, c)
		}
	}
	return mc.removeContainers(stoppedContainers)
}

func (mc MobyClient) RemoveAllContainers() error {
	containers := mc.listContainers()
	return mc.removeContainers(containers)
}

func (mc MobyClient) removeContainer(containerID string) error {
	options := types.ContainerRemoveOptions{Force: true}
	err := mc.client.ContainerRemove(context.Background(), containerID, options)
	return err
}

func (mc MobyClient) removeContainers(containers []types.Container) error {
	removedContainers := []string{}
	for _, c := range containers {
		mc.removeContainer(c.ID)
		shortID := c.ID[:12]
		removedContainers = append(removedContainers, shortID)
	}
	if len(removedContainers) != 0 {
		mc.logger.LogLines(removedContainers)
	}
	return nil
}

func (mc MobyClient) removeImage(imageID string) error {
	options := types.ImageRemoveOptions{Force: true}
	_, err := mc.client.ImageRemove(context.Background(), imageID, options)
	return err
}

func (mc MobyClient) StopContainers() error {
	containers := mc.listContainers()
	for _, c := range containers {
		if c.State == "running" {
			mc.client.ContainerStop(context.Background(), c.ID, 30)
		}
	}
	return nil
}
