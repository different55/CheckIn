// CheckIn is a way to let other people know what the heck you're up to.
//
// It's essentially a status. But like. For your terminal. Has that been
// done already? Has to have been, right?
package main

import (
	"flag"
	"fmt"
	"gitlab.com/golang-commonmark/linkify"
	"html"
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
		GetStatus(flag.Args()[1:])
	case "help":
		fallthrough
	default:
		fmt.Printf("Usage:\n")
		fmt.Printf("\t%v set [--include-wd] [status]\n", os.Args[0])
		fmt.Printf("\t\tstatuses must begin with your username\n")
		fmt.Printf("\t\t--inclue-wd appends your working directory to your status\n")
		fmt.Printf("\t%v get [--freshness=14]\n", os.Args[0])
		fmt.Printf("\t\t--freshness gets all statuses newer than this number of days\n")
		flag.PrintDefaults()
	}
}

// GetStatus collects all recent statuses from all users on the system.
// Any errors that occur while pulling individual statuss are silently ignored
func GetStatus(args []string) {
	getFlags := flag.NewFlagSet(os.Args[0]+" get", flag.ExitOnError)
	freshDays := getFlags.Int("freshness", 14, "get all statuses newer than this number of days")
	htmlOutput := getFlags.Bool("output-html", false, "output statuses as a list of HTML links")
	getFlags.Parse(args)

	freshLimit := time.Now().AddDate(0, 0, *freshDays*-1)

	allStatusPaths, err := filepath.Glob("/home/*/.checkin")
	if err != nil {
		// *Should* never happen, since path is hardcoded and that's the only reason Glob can error out.
		panic(err)
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

	if *htmlOutput {
		fmt.Println("<UL>")
	}

	// Print the contents of all statuses
	for _, statusPath := range freshStatusPaths {
		statusBytes, err := ioutil.ReadFile(statusPath)
		if err != nil {
			continue
		}
		status := string(statusBytes)

		// Get user's name
		// Strip /home from path
		homelessPath, err := filepath.Rel("/home", statusPath)
		if err != nil {
			continue
		}
		username := filepath.Dir(homelessPath)

		// Check to see if file starts with user's name.
		// TODO: split this stuff out into like a normalizeStatus() function.
		if !strings.HasPrefix(status, "~"+username) {
			status = fmt.Sprintf("~%s: %s", username, status)
		}

		// Check to see if file starts with user's name.
		if !strings.HasPrefix(status, "~") {
			// Strip /home from path
			homelessPath, err := filepath.Rel("/home", statusPath)
			if err != nil {
				continue
			}
			// Strip .checkin from path, leaving us with the user's name.
			username = filepath.Dir(homelessPath)
			status = fmt.Sprintf("~%s: %s", username, status)
		}

		// Trim trailing newline (if any)
		status = strings.TrimSpace(status)

		// If HTML output is wanted, detect URLs and replace them with links.
		if *htmlOutput {
			position := 0
			links := linkify.Links(status)
			statusFragments := make([]string, 0, len(links)*2+1)
			for _, v := range links {
				// Copy the stuff before this link and since the last.
				statusFragments = append(statusFragments, status[position:v.Start])

				// Format the link
				link := status[v.Start:v.End]
				if v.Scheme == "" {
					v.Scheme = "http://"
				}

				if !strings.HasPrefix(link, v.Scheme) {
					link = v.Scheme + link
				}

				link = fmt.Sprintf("<A HREF=\"%s\">%s</A>", html.EscapeString(link), html.EscapeString(status[v.Start:v.End]))
				// Copy the actual link
				statusFragments = append(statusFragments, link)
				position = v.End
			}

			// Copy everything that remains
			statusFragments = append(statusFragments, status[position:])

			// Replace the initial ~user with a link to their public_html page.
			homepage := fmt.Sprintf("<A HREF=\"http://tilde.town/~%s\">~%s</A>", username, username)
			statusFragments[0] = strings.Replace(statusFragments[0], "~"+username, homepage, 1)

			status = "<LI>" + strings.Join(statusFragments, "") + "</LI>"
		}

		fmt.Println(status)
	}
	if *htmlOutput {
		fmt.Println("</UL>")
	}
	return
}

// SetStatus sets the curent user's status, either by reading the value from the command line or by prompting the user to input it interactively.
func SetStatus(args []string) error {
	setFlags := flag.NewFlagSet(os.Args[0]+" set", flag.ExitOnError)
	includeWd := setFlags.Bool("include-wd", false, "appends working directory to your message")
	setFlags.Parse(args)

	curUser, err := user.Current()
	if err != nil {
		return err
	}

	// User status is written to ~/.checkin
	outputPath := path.Join(curUser.HomeDir, ".checkin")

	var input string

	// Check if input was provided on the command line.
	if len(setFlags.Args()) > 0 {
		input = ArgsToStatus(setFlags.Args(), curUser)
	} else {
		// Prompt user for input
		input, err = PromptInput()
		if err != nil {
			return err
		}
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
		wd, err := GetFriendlyWd(curUser)
		if err != nil {
			return err
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

// ArgsToStatus concatenates all arguments left over from a call to flag.Parse
// and returns it as a well-formed status.
func ArgsToStatus(args []string, curUser *user.User) string {
	status := strings.Join(args, " ")
	status = strings.TrimRight(status, " \n\t\r")
	if status == "" {
		return status
	}

	if strings.HasPrefix(status, "~"+curUser.Username) {
		return status
	}

	if strings.HasPrefix(status, curUser.Username) {
		return "~" + status
	}

	return fmt.Sprintf("~%s: %s", curUser.Username, status)
}

// GetFriendlyWd returns the friendliest equivalent to the working directory.
// If we're in the user's home directory, we return a path relative to ~user.
// If we're in that user's public_html directory, we return an http link to it.
func GetFriendlyWd(curUser *user.User) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, err
	}

	// Check if we're in the home directory.
	homelessPath, err := filepath.Rel(curUser.HomeDir, wd)

	// If the path couldn't be made relative to the user's home dir at all.
	if err != nil {
		return wd, nil
	}

	// If the path could only be made relative by traversing up the tree, it's also not in the home dir.
	if strings.HasPrefix(homelessPath, "..") {
		return wd, nil
	}

	// If the path is within the user's public_html directory, return an http link.
	weblessPath := strings.TrimPrefix(homelessPath, "public_html")
	if weblessPath != homelessPath {
		return fmt.Sprintf("https://tilde.town/~%s/%s", curUser.Username, weblessPath), nil
	}

	// Otherwise make a friendly ~user/path.
	return filepath.Join("~"+curUser.Username, homelessPath), nil
}
