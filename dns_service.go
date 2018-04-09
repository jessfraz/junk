package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessfraz/k8s-aks-dns-ingress/api/grpc/devops"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const dnsServiceHelp = `Run the dns service.`

func (cmd *dnsServiceCommand) Name() string      { return "dns-service" }
func (cmd *dnsServiceCommand) Args() string      { return "" }
func (cmd *dnsServiceCommand) ShortHelp() string { return dnsServiceHelp }
func (cmd *dnsServiceCommand) LongHelp() string  { return dnsServiceHelp }
func (cmd *dnsServiceCommand) Hidden() bool      { return false }

func (cmd *dnsServiceCommand) Register(fs *flag.FlagSet) {
	// Add our flags.
	fs.StringVar(&cmd.azureconfig, "azureconfig", os.Getenv("AZURE_AUTH_LOCATION"), "Azure service principal configuration file (eg. path to azure.json, defaults to the value of 'AZURE_AUTH_LOCATION' env var")
	fs.StringVar(&cmd.addr, "addr", "0.0.0.0:13377", "TCP address for the grpc server to listen on")
}

type dnsServiceCommand struct {
	azureconfig string
	addr        string
}

func (cmd *dnsServiceCommand) Run(args []string) error {
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logrus.Infof("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	// Get the azure authentication.
	azAuth, err := azure.GetAuthCreds(cmd.azureconfig)
	if err != nil {
		return err
	}

	/*	// Create the Azure provider client.
		dnsClient, err := dns.NewClient(azAuth)
		if err != nil {
			return nil, err
		}*/

	// Create the listener.
	listener, err := net.Listen("tcp", cmd.addr)
	if err != nil {
		return fmt.Errorf("Unable to open TCP socket on %s to listen on: %v", cmd.addr, err)
	}

	// Create the server.
	server := grpc.NewServer()
	devops.RegisterDNSRecordServer(server, devops.NewDNSRecordServer(azAuth))
	logrus.Infof("Starting dns service gRPC server on %s...", cmd.addr)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("Failed to start serving gRPC server on %s: %v", cmd.addr, err)
	}

	return nil
}
