package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	analytics "github.com/google/google-api-go-client/analytics/v3"
	"github.com/jfrazelle/junk/ga/auth"
	"github.com/mitchellh/colorstring"
	"github.com/urfave/cli"
)

const (
	// BANNER is what is printed for help/info output.
	BANNER = `  __ _  __ _
 / _` + "`" + ` |/ _` + "`" + ` |
| (_| | (_| |
 \__, |\__,_|
 |___/
`
	// VERSION is the binary version.
	VERSION = "v0.1.0"

	scope string = "https://www.googleapis.com/auth/analytics.readonly"
)

func doAuth(clientID, secret string, debug bool) (s *analytics.Service, err error) {
	clientID, secret, err = getCreds(clientID, secret, "", "")
	if err != nil {
		return s, err
	}

	a := auth.New(clientID, secret, scope, debug)
	c := a.GetOAuthClient()

	s, err = analytics.New(c)
	if err != nil {
		return s, fmt.Errorf("Creating new analytics service failed: %s", err)
	}
	return s, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "ga"
	app.Version = VERSION
	app.Author = "Jess Frazelle, @frazelledazzell, github.com/jfrazelle"
	app.Usage = "Google Analytics via the Command Line"
	app.EnableBashCompletion = true
	commonFlags := []cli.Flag{
		cli.BoolFlag{Name: "disable-plot", Usage: "Disable plotting"},
		cli.StringFlag{Name: "clientid,c", Value: "", Usage: "Google OAuth Client Id, overrides the .ga-cli files"},
		cli.StringFlag{Name: "secret,s", Value: "", Usage: "Google OAuth Client Secret, overrides the .ga-cli files"},
		cli.BoolFlag{Name: "debug,d", Usage: "Debug mode"},
		cli.BoolFlag{Name: "json", Usage: "Print raw json"},
		cli.BoolFlag{Name: "raw", Usage: "Don't colorize output"},
	}
	app.Flags = commonFlags

	app.Commands = []cli.Command{
		{
			Name:  "accounts",
			Usage: "Get accounts",
			Flags: commonFlags,
			Action: func(c *cli.Context) {
				s, err := doAuth(c.String("clientid"), c.String("secret"), c.Bool("debug"))
				if err != nil {
					printError(err, true)
				}
				accounts, err := getAccounts(s)
				if err != nil {
					printError(err, true)
				}

				// if json
				if c.Bool("json") {
					printPrettyJSON(accounts, c.Bool("raw"))
				} else {
					// print cols
					w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
					fmt.Fprintln(w, "ID\tNAME\tCREATED\tUPDATED")
					for _, account := range accounts {
						fmt.Fprintf(w, "%s\t%s\t%s ago\t%s ago\n", account.Id, account.Name, stringTimeToHuman(account.Created), stringTimeToHuman(account.Updated))
					}
					w.Flush()
				}
			},
		},
		{
			Name:      "configure",
			ShortName: "config",
			Usage:     "Configure your Google API Credentials",
			Flags:     commonFlags,
			Action: func(c *cli.Context) {
				err := configure(c.String("clientid"), c.String("secret"))
				if err != nil {
					printError(err, true)
				}
				fmt.Println(colorstring.Color("[green]ga configured successfully, start running commands"))
			},
		},
		{
			Name:  "now",
			Usage: "Get Realtime Data, dimensions and data reference available at https://developers.google.com/analytics/devguides/reporting/realtime/dimsmets/",
			Flags: append(commonFlags, []cli.Flag{
				cli.StringFlag{Name: "profile,p", Value: "", Usage: "Profile id for which to get data"},
				cli.StringFlag{Name: "metrics,m", Value: "rt:activeUsers", Usage: "Real time metrics to get."},
				cli.StringFlag{Name: "dimensions,dim", Value: "", Usage: "Real time dimensions (comma-separated)"},
				cli.StringFlag{Name: "sort", Value: "", Usage: "Sort to apply"},
			}...),
			Action: func(c *cli.Context) {
				if c.String("profile") != "" {
					s, err := doAuth(c.String("clientid"), c.String("secret"), c.Bool("debug"))
					if err != nil {
						printError(err, true)
					}
					data, err := getNow(s, "ga:"+c.String("profile"), c.String("metrics"), c.String("dimensions"), c.String("sort"))
					if err != nil {
						printError(err, true)
					}
					// if json
					if c.Bool("json") {
						printPrettyJSON(data, c.Bool("raw"))
					} else {
						// print cols
						w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

						// create the header
						var header []string
						for _, h := range data.ColumnHeaders {
							header = append(header, stripColon(h.Name))
						}
						fmt.Fprintln(w, strings.ToUpper(strings.Join(header, "\t")))

						for _, row := range data.Rows {
							fmt.Fprintln(w, strings.Join(row, "\t"))
						}
						w.Flush()

						printTotals(data.TotalsForAllResults)
					}
				} else {
					cli.ShowCommandHelp(c, os.Args[1])
				}
			},
		},
		{
			Name:  "profiles",
			Usage: "Get profiles",
			Flags: append(commonFlags,
				cli.StringFlag{Name: "account", Value: "", Usage: "Account id for which to list profiles"},
			),
			Action: func(c *cli.Context) {
				if c.String("account") != "" {
					s, err := doAuth(c.String("clientid"), c.String("secret"), c.Bool("debug"))
					if err != nil {
						printError(err, true)
					}
					profiles, err := getAccountProfiles(s, c.String("account"))
					if err != nil {
						printError(err, true)
					}

					// if json
					if c.Bool("json") {
						printPrettyJSON(profiles, c.Bool("raw"))
					} else {
						// print cols
						w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
						fmt.Fprintln(w, "ID\tPROPERTY ID\tNAME\tTYPE\tCREATED\tUPDATED")
						for _, profile := range profiles {
							fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s ago\t%s ago\n", profile.Id, profile.WebPropertyId, profile.Name, profile.Type, profile.WebsiteUrl, stringTimeToHuman(profile.Created), stringTimeToHuman(profile.Updated))
						}
						w.Flush()
					}
				} else {
					cli.ShowCommandHelp(c, os.Args[1])
				}
			},
		},
		{
			Name:  "properties",
			Usage: "Get properties",
			Flags: append(commonFlags,
				cli.StringFlag{Name: "account", Value: "", Usage: "Account id for which to list properties"},
			),
			Action: func(c *cli.Context) {
				s, err := doAuth(c.String("clientid"), c.String("secret"), c.Bool("debug"))
				if err != nil {
					printError(err, true)
				}

				var properties []*analytics.Webproperty
				if c.String("account") == "" {
					accounts, err := getAccounts(s)
					if err != nil {
						printError(err, true)
					}
					properties, err = getProperties(s, accounts)
					if err != nil {
						printError(err, true)
					}
				} else {
					properties, err = getPropertiesByAccount(s, c.String("account"))
					if err != nil {
						printError(err, true)
					}
				}

				// if json
				if c.Bool("json") {
					printPrettyJSON(properties, c.Bool("raw"))
				} else {
					// print cols
					w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
					fmt.Fprintln(w, "ID\tACCOUNT ID\tNAME\tURL\tCREATED\tUPDATED\tPROFILE COUNT")
					for _, property := range properties {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s ago\t%s ago\t%v\n", property.Id, property.AccountId, property.Name, property.WebsiteUrl, stringTimeToHuman(property.Created), stringTimeToHuman(property.Updated), property.ProfileCount)
					}
					w.Flush()
				}
			},
		},
	}

	app.Run(os.Args)
}
