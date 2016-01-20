package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jfrazelle/hulk/api/grpc/types"
)

var listCommand = cli.Command{
	Name:  "list",
	Usage: "List jobs",
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) {
		c, err := getClient(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, err := c.ListJobs(context.Background(), &types.ListJobsRequest{})
		if err != nil {
			logrus.Fatalf("ListJobs request failed: %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprint(w, "ID\tNAME\tCMD\tSTATUS\tARTIFACTS\n")
		for _, c := range resp.Jobs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.Id, c.Name, strings.Join(c.Args, " "), c.Status, c.Artifacts)
		}
		w.Flush()
	},
}
