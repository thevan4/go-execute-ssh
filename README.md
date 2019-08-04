# go-execute-ssh [![Go Report Card](https://goreportcard.com/badge/github.com/thevan4/go-execute-ssh)](https://goreportcard.com/report/github.com/thevan4/go-execute-ssh) [![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
Executing commands on remote hosts via ssh

**WARNING: WIP!**

# Example
```package main

import (
	"fmt"
	"log"

	"github.com/thevan4/go-execute-ssh/executessh"
)

func main() {
	host := "1.1.1.1:22"
	user := "admin4eg"
	password := "pass"
	shellPrompt := "#" //or $ for example
	timeoutSeconds := 2

	connection, err := executessh.Connect(host, user, password)
	if err != nil {
		log.Fatal(err)
	}

	output, err := connection.SendCommands(shellPrompt, timeoutSeconds, "show hostname", "show interface mgmt 0 | json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output)
}
