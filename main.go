package main

import (
	"fmt"
	"github.com/jlgrady1/moby/infrastructure"
	"github.com/jlgrady1/moby/interfaces"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
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
)

func main() {
	// log := infrastructure.NewLogger()
	config := infrastructure.LoadConfig()
	mobyClient, _ := interfaces.NewMobyClient()

	kingpin.Version(config.Version)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "quiet":
		mobyClient.Quiet = true
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
		ipAddress, _ := mobyClient.GetIP(*ipContainer)
		fmt.Println(ipAddress)
	}
}
