package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrazelle/ga/analytics"
	"github.com/jfrazelle/ga/auth"
	"os"
)

const (
	VERSION = "v0.1.0"
	BANNER  = `  __ _  __ _
 / _` + "`" + ` |/ _` + "`" + ` |
| (_| | (_| |
 \__, |\__,_|
 |___/
`
	scope string = "https://www.googleapis.com/auth/analytics.readonly"
)

func doAuth(clientId, secret string, debug bool) (s *analytics.Service, err error) {
	clientId, secret, err = getCreds(clientId, secret, "", "")
	if err != nil {
		return s, err
	}

	a := auth.New(clientId, secret, scope, debug)
	c := a.GetOAuthClient()

	s, err = analytics.New(c)
	if err != nil {
		return s, fmt.Errorf("Creating new analytics service failed: %s", err)
	}
	return s, nil
}

func main() {
	app := cli.NewApp()
	app.Name = BANNER
	app.Version = VERSION
	app.Author = "Jess Frazelle, @frazelledazzell, github.com/jfrazelle"
	app.Usage = "Google Analytics via the Command Line"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "disable-plot", Usage: "Disable plotting"},
		cli.StringFlag{Name: "clientid,c", Value: "", Usage: "Google OAuth Client Id, overrides the .ga-cli files"},
		cli.StringFlag{Name: "secret,s", Value: "", Usage: "Google OAuth Client Secret, overrides the .ga-cli files"},
		cli.BoolFlag{Name: "debug,d", Usage: "Debug mode"},
		cli.BoolFlag{Name: "json", Usage: "Print raw json"},
	}

	app.Commands = []cli.Command{
		{
			Name:  "accounts",
			Usage: "Get accounts",
			Action: func(c *cli.Context) {
				s, err := doAuth(c.String("clientid"), c.String("secret"), c.Bool("debug"))
				if err != nil {
					printError(err, true)
				}
				accounts, err := getAccounts(s)
				if err != nil {
					printError(err, true)
				}
				printPrettyJson(accounts, false)
			},
		},
		{
			Name:      "configure",
			ShortName: "config",
			Usage:     "Configure your Google API Credentials",
			Action: func(c *cli.Context) {
				err := configure(c.String("clientid"), c.String("secret"))
				if err != nil {
					printError(err, true)
				}
			},
		},
		{
			Name:  "profiles",
			Usage: "Get profiles",
			Action: func(c *cli.Context) {
				s, err := doAuth(c.String("clientid"), c.String("secret"), c.Bool("debug"))
				if err != nil {
					printError(err, true)
				}
				profiles, err := getAllProfiles(s)
				if err != nil {
					printError(err, true)
				}
				printPrettyJson(profiles, false)
			},
		},
		{
			Name:  "properties",
			Usage: "Get properties",
			Action: func(c *cli.Context) {
				s, err := doAuth(c.String("clientid"), c.String("secret"), c.Bool("debug"))
				if err != nil {
					printError(err, true)
				}
				accounts, err := getAccounts(s)
				if err != nil {
					printError(err, true)
				}
				properties, err := getProperties(s, accounts)
				if err != nil {
					printError(err, true)
				}
				printPrettyJson(properties, false)
			},
		},
	}

	app.Run(os.Args)
}
