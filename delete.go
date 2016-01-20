package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jfrazelle/hulk/api/grpc/types"
	"golang.org/x/net/context"
)

var deleteCommand = cli.Command{
	Name:  "delete",
	Usage: "Delete a job",
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

		logrus.Infof("Job ID: %d", id)

		c, err := getClient(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
		_, err = c.DeleteJob(context.Background(), &types.DeleteJobRequest{
			Id: id,
		})
		if err != nil {
			logrus.Fatalf("DeleteJob request for id %d failed: %v", id, err)
		}
	},
}
