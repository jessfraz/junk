package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Ask(prompt string, output string) (val string, err error) {
	fmt.Printf("%s [%s]: ", prompt, output)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return val, fmt.Errorf("Reading string from prompt failed: %s", err)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return output, nil
	}

	return value, nil
}
