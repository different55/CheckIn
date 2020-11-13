package main

import (
	"fmt"
	"flag"
	"path"
	"os"
	"os/user"
	"io"
	"io/ioutil"
	"bufio"
	"strings"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "set":
		err := SetVenture(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not set: %v", err)
		}		
	default:
		fmt.Printf("Usage:\n\t%v set\n", os.Args[0])
		flag.PrintDefaults()
	}
}

// SetVenture sets the curent user's venture, either by reading the value from the command line or by prompting the user to input it interactively.
func SetVenture(args []string) error {
	curUser, err := user.Current()
	if err != nil {
		return err
	}

	// User status is written to ~/.venture
	outputPath := path.Join(curUser.HomeDir, ".venture")

	// Prompt user for input
	fmt.Printf("What's ~%v been up to?\n~%v", curUser.Username, curUser.Username)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}

	// Remove file on blank input
	if strings.TrimSpace(input) == "" {
		err = os.Remove(outputPath)
		// File already pre-non-existing is considered success.
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}
		return nil
	}

	// Prepend status with user's name
	input = "~"+curUser.Username+input

	// Write file and create if it doesn't exist as world-readable.
	err = ioutil.WriteFile(outputPath, []byte(input), 0644)
	if err != nil {
		return err
	}

	return nil
}