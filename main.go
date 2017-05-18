package main

import (
	"fmt"
	"os"

	"github.com/jlgrady1/moby/infrastructure"
	"github.com/jlgrady1/moby/interfaces"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app     = kingpin.New("moby", "A command-line docker help utility.")
	quiet   = app.Flag("quiet", "Do not print to stdout.").Short('q').Bool()
	logfile = app.Flag("logfile", "File to log to.").Short('f').Default("").String()
	level   = app.Flag("loglevel", "Log Level. Valid values are [TRACE, DEBUG, INFO, WARNING, ERROR]").Short('l').Default("INFO").String()

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

	// test = app.Command("test", "testing")
)

func main() {
	config := infrastructure.LoadConfig()

	kingpin.Version(config.Version)
	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	mobyClient, _ := interfaces.NewMobyClient(*quiet, *logfile)
	var err error
	switch command {

	case removeStopped.FullCommand():
		err = mobyClient.RemoveStoppedContainers()
	case removeAll.FullCommand():
		err = mobyClient.RemoveAllContainers()
	case stopContainers.FullCommand():
		err = mobyClient.StopContainers()
	case cleanImages.FullCommand():
		err = mobyClient.CleanImages()
	case name.FullCommand():
		_, err = mobyClient.GetName(*prefix)
	case ip.FullCommand():
		_, err = mobyClient.GetIP(*ipContainer)
	}

	if err != nil {
		fmt.Println(err)
	}
}
