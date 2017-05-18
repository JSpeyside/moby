package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jlgrady1/moby/infrastructure"
	"github.com/jlgrady1/moby/interfaces"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jlgrady1/go-log"
)

var (
	app   = kingpin.New("moby", "A command-line docker help utility.")
	quiet = app.Flag("quiet", "Do not print to stdout.").Short('q').Bool()

	// Commands
	stopContainers = app.Command("stop-containers", "Stop containers (all by default). Alias: scs").Alias("scs")
	stopContainer  = stopContainers.Arg("container", "Container to stop (all by default).").String()

	removeStopped = app.Command("remove-stopped", "Remove stopped containers. Alias: rms").Alias("rms")
	removeAll     = app.Command("remove-all", "Remove all containers. Alias: rma").Alias("rma")

	cleanImages = app.Command("remove-images", "Remove all untagged images. Alias: rmi").Alias("rmi")

	name   = app.Command("name", "Generates a unique sequential name from a prefix (i.e. 'web' returns 'web-001')")
	prefix = name.Arg("prefix", "Prefix for a new container. For example, issuing web when there are containers web-001 and web-002 would return web-003").Required().String()

	start = app.Command("start", "Starts the default docker machine and configures env")

	ip          = app.Command("ip", "Get the IP for a given container.")
	ipContainer = ip.Arg("name", "The name of the container to fetch the IP from.").Required().String()

	test = app.Command("test", "testing")
)

func main() {
	config := infrastructure.LoadConfig()
	mobyClient, _ := interfaces.NewMobyClient(*quiet)

	kingpin.Version(config.Version)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "quiet":
		fmt.Println("shhh")
	case removeStopped.FullCommand():
		mobyClient.RemoveStoppedContainers()
	case removeAll.FullCommand():
		mobyClient.RemoveAllContainers()
	case stopContainers.FullCommand():
		mobyClient.StopContainers()
	case cleanImages.FullCommand():
		mobyClient.CleanImages()
	case name.FullCommand():
		name, _ := mobyClient.GetName(*prefix)
		fmt.Println(name)
	case ip.FullCommand():
		mobyClient.GetIP(*ipContainer)
	case test.FullCommand():

		// fmt.Println("test")
		log, err := logger.NewLogger("/tmp/moby.log", logger.INFO)
		if err != nil {
			panic(err)
		}
		log.Console("test123")
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			panic(err)
		}

		for _, container := range containers {
			fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		}

	}
}
