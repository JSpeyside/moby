package interfaces

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/jlgrady1/go-log"
	"github.com/jlgrady1/go-utils/format"
	"golang.org/x/net/context"
)

// MobyClient is a client and wrapper around the docker api client
type MobyClient struct {
	client *client.Client
	log    logger.Log
}

//NewMobyClient returns a pointer to a new MobyClient
func NewMobyClient(quiet bool, logfile string) (mobyClient *MobyClient, err error) {
	var log logger.Log
	if quiet == true {
		log = logger.NewMockLogger()
	} else {
		log, err = logger.NewLogger(logfile, logger.TRACE)
		if err != nil {
			return nil, err
		}
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	log.Debug("Created new moby client")
	mobyClient = &MobyClient{
		client: cli,
		log:    log,
	}
	return mobyClient, nil
}

// CleanImages removes all untagged docker images
func (mc *MobyClient) CleanImages() error {
	mc.log.Info("Cleaning images.")
	f := filters.NewArgs()
	report, err := mc.client.ImagesPrune(context.Background(), f)
	if err != nil {
		return err
	}

	mc.log.ConsoleInfo("Cleaned (%d) images reclaming %s", len(report.ImagesDeleted), format.ReadableByteSize(report.SpaceReclaimed))
	return nil
}

// GetIP returns the IP address of a given container
func (mc *MobyClient) GetIP(name string) (string, error) {
	containers, err := mc.listContainersByName(name)
	if err != nil {
		return "", err
	}
	mc.log.Trace("Found (%d) containers by name %s", len(containers), name)
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
		mc.log.ConsoleInfo(bridge.IPAddress)
		return bridge.IPAddress, nil
	}
	return "", nil
}

// GetName returns a sequentially numbered available container name.
// For example, if web-001 exists, GetName "web" would return web-002
func (mc *MobyClient) GetName(name string) (str string, err error) {
	containers, err := mc.listContainersByName(name)
	if err != nil {
		return "", err
	}
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
	newName := fmt.Sprintf("%s-%s", name, sNum)
	mc.log.ConsoleInfo(newName)
	return newName, err
}

func (mc *MobyClient) formatContainers(containers []types.Container) string {
	msg := "[\n"
	for _, c := range containers {
		if c.Names == nil || len(c.Names) < 1 {
			continue
		}
		name := c.Names[0]
		msg = fmt.Sprintf("%s%s,\n", msg, name)
	}
	msg = fmt.Sprintf("%s]", msg)
	return msg
}

func (mc *MobyClient) listImages() []types.ImageInspect {
	// options := types.ImageListOptions{All: true}
	// images, err := mc.client.ImageList(context.Background(), options)
	// if err != nil {
	// 	panic(err)
	// }
	// return images
	return nil
}

func (mc *MobyClient) listContainers() []types.Container {
	options := types.ContainerListOptions{All: true}
	containers, err := mc.client.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}
	return containers
}

func (mc *MobyClient) listContainersByName(name string) ([]types.Container, error) {
	filterArg := filters.NewArgs()
	filterArg.Add("name", name)
	options := types.ContainerListOptions{All: true, Filters: filterArg}
	containers, err := mc.client.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// RemoveStoppedContainers removes all containers with the stopped or exited
// status
func (mc *MobyClient) RemoveStoppedContainers() error {
	mc.log.Info("Removing stopped containers.")
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
func (mc *MobyClient) RemoveAllContainers() error {
	mc.log.Info("Removing all containers.")
	containers := mc.listContainers()
	return mc.removeContainers(containers)
}

func (mc *MobyClient) removeContainer(container *types.Container) error {
	name := container.Names[0]
	mc.log.Trace("Removing container: %s - ", name, container.ID[:12])
	options := types.ContainerRemoveOptions{Force: true}
	err := mc.client.ContainerRemove(context.Background(), container.ID, options)
	return err

}

func (mc *MobyClient) removeContainers(containers []types.Container) error {
	containerString := mc.formatContainers(containers)
	mc.log.Trace("Removing containers %s", containerString)
	removedContainers := []string{}
	for _, c := range containers {
		mc.removeContainer(&c)
		shortID := c.ID[:12]
		removedContainers = append(removedContainers, shortID)
		mc.log.ConsoleInfo(shortID)
	}
	if len(removedContainers) != 0 {
		mc.log.ConsoleInfo("Removed (%d) containers.", len(removedContainers))
	}
	return nil
}

func (mc *MobyClient) removeImage(imageID string) error {
	options := types.ImageRemoveOptions{Force: true}
	_, err := mc.client.ImageRemove(context.Background(), imageID, options)
	return err
}

// StopContainers stops all running containers.
func (mc *MobyClient) StopContainers() error {
	mc.log.Info("Stopping all running containers.")
	containers := mc.listContainers()
	containerStr := mc.formatContainers(containers)
	mc.log.Trace("Found containers %s", containerStr)

	stopCount := 0
	for _, c := range containers {
		if c.State == "running" {
			duration, _ := time.ParseDuration("30s")
			mc.client.ContainerStop(context.Background(), c.ID, &duration)
			stopCount++
		}
	}
	mc.log.ConsoleInfo("Stopped (%d) containers.", stopCount)
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
