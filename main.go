package main

import (
	"fmt"
	"path"
	"os"
	"os/user"
	"io"
	"io/ioutil"
	"bufio"
)

func main() {
	curUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	// User status is written to ~/.venture
	outputPath := path.Join(curUser.HomeDir, ".venture")

	// Prompt user for input
	fmt.Printf("What's ~%v been up to?\n~%v", curUser.Username, curUser.Username)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}

	// Remove file on blank input
	if len(input) < 1 {
		err = os.Remove(outputPath)
		if err != nil {
			panic(err)
		}
		return
	}

	// Prepend status with user's name
	input = "~"+curUser.Username+input

	// Write file and create if it doesn't exist as world-readable.
	err = ioutil.WriteFile(outputPath, []byte(input), 0644)
	if err != nil {
		panic(err)
	}
}