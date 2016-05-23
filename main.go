package main

import (
	"github.com/jlgrady1/moby/infrastructure"
	"github.com/jlgrady1/moby/interfaces"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	app   = kingpin.New("moby", "A command-line docker help utility.")
	quiet = app.Flag("quiet", "Do not print to stdout.").Short('q').Bool()

	// Commands
	stopContainers = app.Command("stop-containers", "Stop containers (all by default).").Alias("scs")
	stopContainer  = stopContainers.Arg("container", "Container to stop (all by default).").String()

	removeStopped = app.Command("remove-stopped", "Remove stopped containers").Alias("rms")
	removeAll     = app.Command("remove-all", "Remove all containers").Alias("rma")

	cleanImages = app.Command("remove-images", "Remove all untagged images.").Alias("rmi")

	name = app.Command("name", "Generates a unique sequential name from a prefix (i.e. 'web' returns 'web-001')")

	start = app.Command("start", "Starts the default docker machine and configures env")

	ip = app.Command("ip", "Get the IP for a given container.")
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
	}
}
