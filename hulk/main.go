package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = ` _           _ _
| |__  _   _| | | __
| '_ \| | | | | |/ /
| | | | |_| | |   <
|_| |_|\__,_|_|_|\_\

 Ultimate job/build runner -OR- bash execution as a service.
 Version: %s

`
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

func main() {
	app := cli.NewApp()
	app.Name = "hulk"
	app.Version = VERSION
	app.Author = "@jessfraz"
	app.Email = "no-reply@butts.com"
	app.Usage = "Ultimate job/build queue runner and scheduler."
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if ctx.GlobalString("addr") == "" {
			return fmt.Errorf("Address for GRPC API to listen cannot be empty")
		}
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr, a",
			Value: "/run/hulk/hulk.sock",
			Usage: "Address on which GRPC API will listen",
		},
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
