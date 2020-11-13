package main

import (
	"fmt"
	"flag"
	"path"
	"path/filepath"
	"os"
	"os/user"
	"io"
	"io/ioutil"
	"bufio"
	"strings"
	"time"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "set":
		err := SetVenture(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not set: %v", err)
		}
	case "get":
		err := GetVenture(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get: %v", err)
		}
	default:
		fmt.Printf("Usage:\n\t%v <set|get>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

// GetVenture collects all recent ventures from all users on the system.
// Any errors that occur while pulling individual ventures are silently ignored
func GetVenture(args []string) error {
	// Set the expiration date for ventures at 2 weeks.
	freshLimit := time.Now().AddDate(0, 0, -14)

	allVenturePaths, err := filepath.Glob("/home/*/.venture")
	if err != nil {
		// *Should* never happen, since path is hardcoded and that's the only reason Glob can error out.
		return err
	}

	// Filter out any ventures that are older than our cutoff time.
	freshVenturePaths := make([]string, 0, len(allVenturePaths))
	for _, venturePath := range allVenturePaths {
		ventureInfo, err := os.Stat(venturePath)
		if err != nil {
			continue
		}
		if ventureInfo.ModTime().After(freshLimit) {
			freshVenturePaths = append(freshVenturePaths, venturePath)
		}
	}

	// Print the contents of all ventures
	for _, venturePath := range freshVenturePaths {
		ventureBytes, err := ioutil.ReadFile(venturePath)
		if err != nil {
			continue
		}
		venture := string(ventureBytes)

		// Check to see if file starts with user's name.
		if !strings.HasPrefix(venture, "~") {
			// Strip /home from path
			homelessPath, err := filepath.Rel("/home", venturePath)
			if err != nil {
				continue
			}
			// Strip .venture from path, leaving us with the user's name.
			username := filepath.Dir(homelessPath)
			venture = fmt.Sprintf("~%s: %s", username, venture)
		}

		// Trim trailing newline (if any)
		venture = strings.TrimSpace(venture)
		
		fmt.Println(venture)
	}
	return nil	
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