package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = ` _           _ _
| |__  _   _| | | __
| '_ \| | | | | |/ /
| | | | |_| | |   <
|_| |_|\__,_|_|_|\_\

 Ultimate job/build queue runner and scheduler.
 Version: %s

`
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

func main() {
	app := cli.NewApp()
	app.Name = "hulk"
	app.Version = VERSION
	app.Author = "@jfrazelle"
	app.Email = "no-reply@butts.com"
	app.Usage = "Ultimate job/build queue runner and scheduler."
	app.Before = func(context *cli.Context) error {
		if context.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "run in debug mode",
		},
	}
	app.Commands = []cli.Command{
		deleteCommand,
		listCommand,
		logsCommand,
		serverCommand,
		startCommand,
		statusCommand,
	}
	app.Run(os.Args)
}
