package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "set":
		err := SetStatus(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not set: %v", err)
		}
	case "get":
		err := GetStatus(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get: %v", err)
		}
	default:
		fmt.Printf("Usage:\n\t%v <set|get> [--help]\n", os.Args[0])
		flag.PrintDefaults()
	}
}

// GetStatus collects all recent statuses from all users on the system.
// Any errors that occur while pulling individual statuss are silently ignored
func GetStatus(args []string) error {
	getFlags := flag.NewFlagSet(os.Args[0]+" get", flag.ExitOnError)
	freshDays := getFlags.Int("freshness", 14, "get all statuses newer than this number of days")
	getFlags.Parse(args)

	freshLimit := time.Now().AddDate(0, 0, *freshDays*-1)

	allStatusPaths, err := filepath.Glob("/home/*/.checkin")
	if err != nil {
		// *Should* never happen, since path is hardcoded and that's the only reason Glob can error out.
		return err
	}

	// Filter out any statuses that are older than our cutoff time.
	freshStatusPaths := make([]string, 0, len(allStatusPaths))
	for _, statusPath := range allStatusPaths {
		statusInfo, err := os.Stat(statusPath)
		if err != nil {
			continue
		}
		if statusInfo.ModTime().After(freshLimit) {
			freshStatusPaths = append(freshStatusPaths, statusPath)
		}
	}

	// Print the contents of all statuses
	for _, statusPath := range freshStatusPaths {
		statusBytes, err := ioutil.ReadFile(statusPath)
		if err != nil {
			continue
		}
		status := string(statusBytes)

		// Check to see if file starts with user's name.
		if !strings.HasPrefix(status, "~") {
			// Strip /home from path
			homelessPath, err := filepath.Rel("/home", statusPath)
			if err != nil {
				continue
			}
			// Strip .checkin from path, leaving us with the user's name.
			username := filepath.Dir(homelessPath)
			status = fmt.Sprintf("~%s: %s", username, status)
		}

		// Trim trailing newline (if any)
		status = strings.TrimSpace(status)

		fmt.Println(status)
	}
	return nil
}

// SetStatus sets the curent user's status, either by reading the value from the command line or by prompting the user to input it interactively.
func SetStatus(args []string) error {
	setFlags := flag.NewFlagSet(os.Args[0]+" set", flag.ExitOnError)
	includeWd := setFlags.Bool("include-wd", false, "if set, appends working directory to your message")
	setFlags.Parse(args)

	curUser, err := user.Current()
	if err != nil {
		return err
	}

	// User status is written to ~/.checkin
	outputPath := path.Join(curUser.HomeDir, ".checkin")

	// Prompt user for input
	input, err := PromptInput()
	if err != nil {
		return err
	}

	// Remove file on blank input
	if input == "" {
		err = os.Remove(outputPath)
		// File already pre-non-existing is considered success.
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}
		return nil
	}

	if *includeWd {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		// If they're in their home directory, swap /home/user out for ~user
		homelessPath, err := filepath.Rel(curUser.HomeDir, wd)

		// Gotta be careful because Rel doesn't care if wd is a subdirectory of HomeDir,
		// it'll slather as much "../../.." as it needs to make it relative.
		if err == nil && !strings.HasPrefix(homelessPath, "..") {
			wd = filepath.Join("~"+curUser.Username, homelessPath)
		}

		input = fmt.Sprintf("%s (%s)", input, wd)
	}

	input = input + "\n"

	// Write file and create if it doesn't exist as world-readable.
	err = ioutil.WriteFile(outputPath, []byte(input), 0644)
	if err != nil {
		return err
	}

	return nil
}
