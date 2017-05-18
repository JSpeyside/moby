package interfaces

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/jlgrady1/moby/infrastructure"
	"golang.org/x/net/context"
)

// MobyClient is a client and wrapper around the docker api client
type MobyClient struct {
	client *client.Client
	logger *infrastructure.Logger
	Quiet  bool
}

//NewMobyClient returns a pointer to a new MobyClient
func NewMobyClient() (*MobyClient, error) {
	var cli *client.Client
	var err error
	systemOS := runtime.GOOS
	if systemOS == "darwin" || systemOS == "linux" {
		// cli, err = client.NewEnvClient()
		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		cli, err = client.NewClient("unix:///var/run/docker.sock", "v1.23", nil, defaultHeaders)
	} else {
		// defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		// cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
		panic(fmt.Sprintf("Unsupported OS `%s`. Only Mac OS X is currently supported", systemOS))
	}
	logger, err := infrastructure.NewLogger(false)
	mobyClient := &MobyClient{cli, logger, false}
	return mobyClient, err
}

// CleanImages removes all untagged docker images
func (mc MobyClient) CleanImages() error {
	var err error
	previousCount := -1
	removeCount := 0
	for removeCount != previousCount {
		previousCount = removeCount
		removeCount, err = mc.cleanTags()
		if err != nil {
			return err
		}
	}
	mc.logger.Log("Cleaned images.")
	return nil
}

func (mc MobyClient) cleanTags() (int, error) {
	images := mc.listImages()
	removeCount := 0

	for _, i := range images {
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
			removeCount++
		}
	}
	return removeCount, nil
}

// GetIP returns the IP address of a given container
func (mc MobyClient) GetIP(name string) (string, error) {
	containers := mc.listContainersByName(name)
	for _, c := range containers {
		cName := c.Names[0]
		if len(cName) > 0 {
			cName = cName[1:len(cName)]
		}
		if name != cName {
			continue
		}
		bridge, ok := c.NetworkSettings.Networks["bridge"]
		if !ok {
			continue
		}
		return bridge.IPAddress, nil
	}
	return "", nil
}

// GetName returns a sequentially numbered available container name.
// For example, if web-001 exists, GetName "web" would return web-002
func (mc MobyClient) GetName(name string) (str string, err error) {
	containers := mc.listContainersByName(name)
	num := 0
	for _, c := range containers {
		cName := c.Names[0]
		if len(cName) > 0 {
			cName = cName[1:len(cName)]
		}
		parts := strings.SplitN(cName, "-", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] != name {
			continue
		}
		strNum := parts[1]
		dnum, strerr := strconv.Atoi(strNum)
		if strerr != nil {
			continue
		}
		if dnum > num {
			num = dnum
		}
	}
	num++
	sNum, _ := padNum(num)
	return fmt.Sprintf("%s-%s", name, sNum), err
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

func (mc MobyClient) listContainersByName(name string) []types.Container {
	filterArg := filters.NewArgs()
	filterArg.Add("name", name)
	options := types.ContainerListOptions{All: true, Filter: filterArg}
	containers, err := mc.client.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}
	return containers
}

// RemoveStoppedContainers removes all containers with the stopped or exited
// status
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

// RemoveAllContainers removes all existing containers. Use with caution!
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
		msg := fmt.Sprintf("Removed (%d) containers.", len(removedContainers))
		mc.logger.Log(msg)
	}
	return nil
}

func (mc MobyClient) removeImage(imageID string) error {
	options := types.ImageRemoveOptions{Force: true}
	_, err := mc.client.ImageRemove(context.Background(), imageID, options)
	return err
}

// StopContainers stops all running containers.
func (mc MobyClient) StopContainers() error {
	containers := mc.listContainers()
	stopCount := 0
	for _, c := range containers {
		if c.State == "running" {
			duration, _ := time.ParseDuration("30s")
			mc.client.ContainerStop(context.Background(), c.ID, &duration)
			stopCount++
		}
	}
	msg := fmt.Sprintf("Stopped (%d) containers.", stopCount)
	mc.logger.Log(msg)
	return nil
}

func padNum(num int) (string, error) {
	if num > 999 {
		return "", fmt.Errorf("Number %d is too large. Only numbers up to 999 are supported", num)
	} else if num > 100 {
		return fmt.Sprintf("%d", num), nil
	} else if num > 10 {
		return fmt.Sprintf("0%d", num), nil
	}
	return fmt.Sprintf("00%d", num), nil
}
