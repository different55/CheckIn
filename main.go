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

	outputPath := path.Join(curUser.HomeDir, ".venture")

	fmt.Printf("What's ~%v been up to?\n~%v", curUser.Username, curUser.Username)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}

	if len(input) < 1 {
		err = os.Remove(outputPath)
		if err != nil {
			panic(err)
		}
		return
	}

	input = "~"+curUser.Username+input
	
	err = ioutil.WriteFile(outputPath, []byte(input), 0644)
	if err != nil {
		panic(err)
	}
}