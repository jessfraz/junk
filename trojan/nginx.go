package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
	"github.com/satori/go.uuid"
)

type route struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	ServerName string `json:"server_name"`
	SSL        bool   `json:"ssl"`
	Auth       bool   `json:"auth"`
}

func setRoutes(e *etcd.Client) error {
	var (
		data         []byte
		ipsChanged   bool
		portsChanged bool
		routes       []route
	)

	// fetch routes from etcd
	resp, err := e.Get("/nginx/routes", false, false)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp.Node.Value), &routes); err != nil {
		return fmt.Errorf("Unmarshaling routes from etcd failed: %v", err)
	}

	// get the current ips and ports
	ips, ports, err := getCurrentFirewall(e)
	if err != nil {
		return nil
	}

	for _, r := range routes {
		// check if we need to add the ip to ips
		if r.IP != "0.0.0.0" && r.IP != "127.0.0.1" {
			// iterate through the current ips to see if we have this one
			ipfound := false
			for _, ip := range ips {
				if ip == r.IP {
					ipfound = true
					break
				}
			}
			if !ipfound {
				// append it to ips
				ips = append(ips, r.IP)
				ipsChanged = true
			}
		}

		// check if we need to add the port to ports
		portfound := false
		for _, p := range ports {
			if p.Port == r.Port && p.Proto == "tcp" {
				portfound = true
				break
			}
		}

		if !portfound {
			// append it to ports
			ports = append(ports, port{
				Proto: "tcp",
				Port:  r.Port,
			})
			portsChanged = true
		}

		// create a uid for upstream
		uid := uuid.NewV4()

		// create the .conf file
		content := template
		if r.SSL {
			content = sslTemplate
		}
		content = strings.Replace(content, "{SERVERNAME}", r.ServerName, -1)
		content = strings.Replace(content, "{NAME}", uid.String(), -1)
		content = strings.Replace(content, "{IP}", r.IP, -1)
		content = strings.Replace(content, "{PORT}", strconv.Itoa(r.Port), -1)
		if r.Auth {
			auth := `auth_basic "Restricted";
auth_basic_user_file ` + path.Join(nginxDir, ".htpasswd") + `;
`
			content = strings.Replace(content, "#{AUTH}", auth, -1)
		}
		if err := ioutil.WriteFile(path.Join(nginxDir, "conf.d", r.ServerName+".conf"), []byte(content), 0755); err != nil {
			return err
		}
	}

	// if the ips changed, update the value
	if ipsChanged {
		data, err = json.Marshal(ips)
		if err != nil {
			return fmt.Errorf("Marshalling ips failed: %v", err)
		}
		if _, err = e.Set("/firewall/ips", string(data), 0); err != nil {
			return err
		}
	}

	// if the ports changed, update the value
	if portsChanged {
		data, err = json.Marshal(ports)
		if err != nil {
			return fmt.Errorf("Marshalling ports failed: %v", ports)
		}
		if _, err = e.Set("/firewall/ports", string(data), 0); err != nil {
			return err
		}
	}

	return nil
}

func initializeNginx(e *etcd.Client) error {
	// make the directory if it doesn't exist
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		return err
	}

	// setup default nginx.conf if it does not exist
	nginxConfPath := path.Join(nginxDir, "nginx.conf")
	if _, err := os.Stat(nginxConfPath); os.IsNotExist(err) {
		// write the default nginx conf
		content := nginxConf
		if sslCrt != "" && sslKey != "" {
			content = strings.Replace(strings.Replace(sslOptions, "{PATH_TO_SERVER_CRT}", sslCrt, 1), "{PATH_TO_SERVER_KEY}", sslKey, 1)
			content = strings.Replace(nginxConf, "#{InsertSSLHere}", content, 1)
		}
		if err := ioutil.WriteFile(nginxConfPath, []byte(content), 0755); err != nil {
			return err
		}

		// write the mime.types file
		if err := ioutil.WriteFile(path.Join(nginxDir, "mime.types"), []byte(mimeTypes), 0755); err != nil {
			return err
		}
	}

	// make the config directory
	// if it doesn't exist
	if err := os.MkdirAll(path.Join(nginxDir, "conf.d"), 0755); err != nil {
		return err
	}

	// intialize the additional conf files

	return setRoutes(e)
}

func nginxLoop(e *etcd.Client, update chan *etcd.Response) {
	for resp := range update {
		logrus.Infof("Processing updated conf for %s", resp.Node.Key)

		if err := setRoutes(e); err != nil {
			logrus.Warnf("Setting routes failed: %v", err)
		}
	}
}
