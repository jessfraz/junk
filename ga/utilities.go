package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/mitchellh/colorstring"
)

func stringTimeToHuman(ts string) (ds string) {
	tt, err := time.Parse("2006-01-02T15:04:05Z07:00", ts)
	if err != nil {
		return ts
	}
	d := time.Since(tt)
	return units.HumanDuration(d)
}

func printPrettyJSON(v interface{}, printRaw bool) {
	b, _ := json.MarshalIndent(v, "", "  ")
	code := string(b)

	if printRaw {
		fmt.Println(code)
	} else {
		// look for Pygments to highlight code
		var pygmentsBin = "pygmentize"
		if _, err := exec.LookPath(pygmentsBin); err != nil {
			printWarning(errors.New("Could not find `Pygments` installed on your system. To pretty print the json please pip install `Pygments` and make sure it is in your path."))
			fmt.Println(code)
			return
		}

		var out bytes.Buffer
		var stderr bytes.Buffer

		cmd := exec.Command(pygmentsBin, "-fterminal256", "-Ofull,style=native", "-ljson")
		cmd.Stdin = strings.NewReader(code)
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			printError(fmt.Errorf("Error running pygmentize: %s", stderr.String()), false)
			fmt.Println(code)
			return
		}
		fmt.Println(out.String())
	}

	return
}

func stripColon(s string) string {
	colI := strings.Index(s, ":")
	if colI > -1 && colI < len(s)-1 {
		return s[colI+1:]
	}
	return s
}

func printTotals(totals map[string]string) {
	if len(totals) > 0 {
		header := []string{"Totals"}
		row := []string{"-"}

		fmt.Println("")
		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		for k, v := range totals {
			header = append(header, stripColon(k))
			row = append(row, v)
		}

		fmt.Fprintln(w, colorstring.Color("[cyan]"+strings.Join(header, "\t")))
		fmt.Fprintln(w, colorstring.Color("[blue]"+strings.Join(row, "\t")))

		w.Flush()
	}
	return
}

func printWarning(err error) {
	fmt.Println(colorstring.Color("[yellow]" + err.Error()))
}

func printError(err error, exit bool) {
	fmt.Println(colorstring.Color("[red]" + err.Error()))
	if exit {
		os.Exit(1)
	}
}
