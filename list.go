package main

import "github.com/codegangsta/cli"

var listCommand = cli.Command{
	Name:  "list",
	Usage: "List jobs",
	Flags: []cli.Flag{},
	Action: func(context *cli.Context) {
	},
}
