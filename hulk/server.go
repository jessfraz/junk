package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jfrazelle/junk/hulk/api/grpc/server"
	"github.com/jfrazelle/junk/hulk/api/grpc/types"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var serverCommand = cli.Command{
	Name:    "server",
	Aliases: []string{"daemon"},
	Usage:   "Start hulk server",
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
		cli.StringFlag{
			Name:  "smtp-server",
			Usage: "SMTP server for email notifications",
		},
		cli.StringFlag{
			Name:  "smtp-sender",
			Usage: "SMTP default sender email address for email notifications",
		},
		cli.StringFlag{
			Name:  "smtp-username",
			Usage: "SMTP server username",
		},
		cli.StringFlag{
			Name:  "smtp-pass",
			Usage: "SMTP server password",
		},
	},
	Action: func(ctx *cli.Context) {
		if err := startServer(
			ctx.GlobalString("addr"),
			ctx.String("artifacts-dir"),
			ctx.String("state-dir"),
			ctx.String("smtp-server"),
			ctx.String("smtp-sender"),
			ctx.String("smtp-username"),
			ctx.String("smtp-pass"),
		); err != nil {
			logrus.Fatal(err)
		}
	},
}

func startServer(address, artifactsDir, stateDir, smtpServer, smtpSender, smtpUsername, smtpPassword string) error {
	if err := os.RemoveAll(address); err != nil {
		return fmt.Errorf("attempt to remove %s failed: %v", address, err)
	}

	l, err := net.Listen("unix", address)
	if err != nil {
		return fmt.Errorf("starting listener at %s failed: %v", address, err)
	}

	s := grpc.NewServer()
	svr, err := server.NewServer(artifactsDir, stateDir, smtpServer, smtpSender, smtpUsername, smtpPassword)
	if err != nil {
		return fmt.Errorf("Creating new server failed: %v", err)
	}

	types.RegisterAPIServer(s, svr)
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
