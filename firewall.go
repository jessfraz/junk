package main

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
	"github.com/jfrazelle/trojan/iptables"
)

type port struct {
	Port  int    `json:"port"`
	Proto string `json:"proto"`
}

func getCurrentFirewall(e *etcd.Client) (ips []string, ports []port, err error) {
	// fetch ips from etcd
	resp, err := e.Get("/firewall/ips", false, false)
	if err != nil {
		return ips, ports, err
	}
	if err := json.Unmarshal([]byte(resp.Node.Value), &ips); err != nil {
		return ips, ports, fmt.Errorf("Unmarshaling ips from etcd failed: %v", err)
	}

	// fetch ports from etcd
	resp, err = e.Get("/firewall/ports", false, false)
	if err != nil {
		return ips, ports, err
	}
	if err := json.Unmarshal([]byte(resp.Node.Value), &ports); err != nil {
		return ips, ports, fmt.Errorf("Unmarshalling ports from etcd failed: %v", err)
	}

	return ips, ports, nil
}

func setIPRules(e *etcd.Client) error {
	var chain string = "INPUT"

	ips, ports, err := getCurrentFirewall(e)
	if err != nil {
		return nil
	}

	// flush the existing rules
	if _, err := iptables.Raw("-F", chain); err != nil {
		return fmt.Errorf("Flusing iptables chain %q failed: %v", chain, err)
	}
	log.Debugf("Flushed iptables chain %q", chain)

	// apply new rules
	for _, p := range ports {
		if err := applyPortRules(ips, p.Port, p.Proto); err != nil {
			return err
		}
	}

	log.Infof("Updated rules in iptables chain %q", chain)
	return nil
}

// iptables -A INPUT -i eth0 -p tcp --dport 8080 -j DROP
// iptables -I INPUT -i eth0 -s 127.0.0.1 -p tcp --dport 8080 -j ACCEPT
func applyPortRules(ips []string, port int, proto string) error {
	// process the DROP rule
	if _, err := iptables.Raw("-A", "INPUT", "-i", iface, "-p", proto, "--dport", fmt.Sprint(port), "-j", "DROP"); err != nil {
		return err
	}

	for _, a := range ips {
		if _, err := iptables.Raw("-I", "INPUT", "-i", iface, "-s", a, "-p", proto, "--dport", fmt.Sprint(port), "-j", "ACCEPT"); err != nil {
			return err
		}
	}

	return nil
}

func firewallLoop(e *etcd.Client, update chan *etcd.Response) {
	for resp := range update {
		log.Infof("Processing updated rules for %s", resp.Node.Key)

		if err := setIPRules(e); err != nil {
			log.Warnf("Updating IP rules failed: %v", err)
		}
	}
}
