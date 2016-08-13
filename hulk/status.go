package main

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jfrazelle/hulk/api/grpc/types"
)

var statusCommand = cli.Command{
	Name:  "status",
	Usage: "Get job status",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "id",
			Usage: "Job ID",
		},
	},
	Action: func(ctx *cli.Context) {
		id := uint32(ctx.Int("id"))
		if id <= 0 {
			cli.ShowSubcommandHelp(ctx)
			logrus.Fatalf("Pass a job ID.")
		}

		c, err := getClient(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, err := c.State(context.Background(), &types.StateRequest{
			Id: id,
		})
		if err != nil {
			logrus.Fatalf("Status request for id %d failed: %v", id, err)
		}
		fmt.Println(resp.Status)
	},
}
