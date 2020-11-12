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

	fmt.Printf("What's ~%v been up to?\n~%v", curUser.Username, curUser.Username)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}

	input = "~"+curUser.Username+input
	
	outputPath := path.Join(curUser.HomeDir, ".venture")
	err = ioutil.WriteFile(outputPath, []byte(input), 0644)
	if err != nil {
		panic(err)
	}
}