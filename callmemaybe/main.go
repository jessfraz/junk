package main

import (
	"fmt"
	"syscall"
)

func main() {
	// Setenv
	if err := syscall.Setuid(0); err == nil {
		fmt.Println(`Setuid(0) did not fail wahhhh`)
	}

	// Unlink
	if err := syscall.Unlink("/fixtures/call"); err == nil {
		fmt.Println(`Unlink("/fixtures/call") did not fail, wahhhh`)
	}
}
