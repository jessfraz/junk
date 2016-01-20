package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var logsCommand = cli.Command{
	Name:  "logs",
	Usage: "Stream job logs",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "id",
			Usage: "Job ID",
		},
	},
	Action: func(context *cli.Context) {
		if context.Int("id") <= 0 {
			logrus.Fatalf("Pass a job ID.")
		}
		logrus.Infof("Job ID: %d", context.Int("id"))
	},
}
