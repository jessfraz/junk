package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
)

const (
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

var (
	iface    string
	nginxDir string
	sslCrt   string
	sslKey   string
	debug    bool
	version  bool
)

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&iface, "iface", "eth0", "default firewall interface")
	flag.StringVar(&nginxDir, "nginx-dir", "/etc/nginx", "path to nginx directory")
	flag.StringVar(&sslCrt, "ssl-crt", "", "if using ssl, path to ssl certificate")
	flag.StringVar(&sslKey, "ssl-key", "", "if using ssl, path to ssl key")
	flag.Parse()
}

func main() {
	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	machine := os.Getenv("ETCD")
	if machine == "" {
		machine = "http://0.0.0.0:4001"
	}
	logrus.Infof("Connecting to %s", machine)

	// initialize etcd client
	e := etcd.NewClient([]string{machine})

	// get the interface values
	// typical/default is eth0
	resp, err := e.Get("/firewall/interface", false, false)
	if err != nil {
		logrus.Fatal(err)
	}

	if resp.Node.Value == "" {
		logrus.Debug("/firewall/interface returned empty contents, using %q", iface)

		if _, err = e.Set("/firewall/interface", iface, 0); err != nil {
			logrus.Fatal(err)
		}
	} else {
		iface = resp.Node.Value
	}

	// initialize values for firewall ip rules
	if err := setIPRules(e); err != nil {
		logrus.Fatal(err)
	}

	// watch for changes to the firewall
	updateFirewall := make(chan *etcd.Response)
	go firewallLoop(e, updateFirewall)
	if _, err := e.Watch("/firewall", 0, true, updateFirewall, nil); err != nil {
		logrus.Error(err)
	}

	// watch for changes to nginx
	// nginx changes will trigger a change in firewall
	//updateNginx := make(chan *etcd.Response)
	//go nginxLoop(e, updateNginx)
	//if _, err := e.Watch("/nginx", 0, true, updateNginx, nil); err != nil {
	//log.Error(err)
	//}
}
