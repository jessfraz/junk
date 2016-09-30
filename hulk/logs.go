package main

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/jessfraz/junk/hulk/api/grpc/types"
	"github.com/urfave/cli"
)

var logsCommand = cli.Command{
	Name:  "logs",
	Usage: "Stream job logs",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "id",
			Usage: "Job ID",
		},
		cli.BoolFlag{
			Name:  "follow, f",
			Usage: "Follow log output",
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
			Id:     id,
			Follow: ctx.Bool("follow"),
		})
		if err != nil {
			logrus.Fatalf("Logs request for id %d failed: %v", id, err)
		}
		for {
			l, err := logs.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				logrus.Fatalf("Receiving logs for id %d failed: %v", id, err)
			}
			fmt.Println(l.Log)
		}
	},
}
