package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/jessfraz/junk/hulk/api/grpc/types"
	"github.com/urfave/cli"
)

var startCommand = cli.Command{
	Name:    "start",
	Aliases: []string{"run"},
	Usage:   "Start a job",
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
		cli.StringFlag{
			Name:  "email",
			Usage: "Email address to send job results",
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

		c, err := getClient(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, err := c.StartJob(context.Background(), &types.StartJobRequest{
			Name:           name,
			Args:           args,
			Artifacts:      artifacts,
			EmailRecipient: ctx.String("email"),
		})
		if err != nil {
			logrus.Fatalf("StartJob request for name %s failed: %v", name, err)
		}
		fmt.Printf("%d\n", resp.Id)
	},
}
