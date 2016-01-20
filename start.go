package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var startCommand = cli.Command{
	Name:  "start",
	Usage: "Start a job",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Usage: "Job name",
		},
		cli.StringFlag{
			Name:  "artifacts",
			Usage: "Where artifacts from a job are stored, relative to the temp dir where job is run",
		},
		cli.StringFlag{
			Name:  "cmd, c",
			Usage: "Command to run",
		},
	},
	Action: func(context *cli.Context) {
		if context.String("name") == "" || context.String("cmd") == "" {
			logrus.Fatalf("Pass a job name and command.")
		}
		logrus.Infof("Job Name: %s", context.String("name"))
		logrus.Infof("Job Command: %s", context.String("command"))
	},
}
