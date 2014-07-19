package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/colorstring"
	"os"
	"os/exec"
	"strings"
)

func printPrettyJson(v interface{}, printRaw bool) {
	json_byte, _ := json.MarshalIndent(v, "", "  ")
	code := string(json_byte)

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

func printWarning(err error) {
	fmt.Println(colorstring.Color("[yellow]" + err.Error()))
}

func printError(err error, exit bool) {
	fmt.Println(colorstring.Color("[red]" + err.Error()))
	if exit {
		os.Exit(1)
	}
}
