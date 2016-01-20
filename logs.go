package main

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jfrazelle/hulk/api/grpc/types"
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
		logs, err := c.Logs(context.Background(), &types.LogsRequest{
			Id: id,
		})
		if err != nil {
			logrus.Fatalf("Logs request for id %d failed: %v", id, err)
		}
		for {
			l, err := logs.Recv()
			if err != nil {
				logrus.Fatalf("Receiving logs for id %d failed: %v", id, err)
			}
			fmt.Println(l.Log)
		}
	},
}
