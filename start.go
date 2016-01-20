package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jfrazelle/hulk/api/grpc/types"
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
	Action: func(ctx *cli.Context) {
		name := ctx.String("name")
		args := strings.Split(ctx.String("cmd"), " ")
		artifacts := ctx.String("artifacts")

		if name == "" || len(args) <= 0 {
			cli.ShowSubcommandHelp(ctx)
			logrus.Fatalf("Pass a job name and command.")
		}

		logrus.Infof("Job name: %s", name)
		logrus.Infof("Job args: %#v", args)

		c, err := getClient(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, err := c.StartJob(context.Background(), &types.StartJobRequest{
			Name:      name,
			Args:      args,
			Artifacts: artifacts,
		})
		if err != nil {
			logrus.Fatalf("StartJob request for name %s failed: %v", name, err)
		}
		fmt.Printf("%d\n", resp.Id)
	},
}
