package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jfrazelle/hulk/api/grpc/server"
	"github.com/jfrazelle/hulk/api/grpc/types"
	"google.golang.org/grpc"
)

var serverCommand = cli.Command{
	Name:  "server",
	Usage: "Start hulk server",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "artifacts-dir",
			Value: "/var/lib/hulk",
			Usage: "Artifacts directory for saving the artifacts from the jobs",
		},
		cli.StringFlag{
			Name:  "state-dir",
			Value: "/run/hulk",
			Usage: "State directory",
		},
	},
	Action: func(ctx *cli.Context) {
		if err := startServer(
			ctx.GlobalString("addr"),
			ctx.String("state-dir"),
			ctx.String("artifacts-dir"),
		); err != nil {
			logrus.Fatal(err)
		}
	},
}

func startServer(address, artifactsDir, stateDir string) error {
	if err := os.RemoveAll(address); err != nil {
		return fmt.Errorf("attempt to remove %s failed: %v", address, err)
	}
	l, err := net.Listen("unix", address)
	if err != nil {
		return fmt.Errorf("starting listener at %s failed: %v", address, err)
	}
	s := grpc.NewServer()
	types.RegisterAPIServer(s, server.NewServer(artifactsDir, stateDir))
	logrus.Debugf("GRPC API listen on %s", address)
	return s.Serve(l)
}

func getClient(ctx *cli.Context) (types.APIClient, error) {
	address := ctx.GlobalString("addr")
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	dialOpts = append(dialOpts,
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		},
		))
	conn, err := grpc.Dial(address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating connection to %s failed: %v", address, err)
	}
	return types.NewAPIClient(conn), nil
}
