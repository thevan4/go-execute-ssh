# go-execute-ssh
Executing ssh commands on remote hosts

# Example
```package main

import (
	"fmt"
	"log"

	"./executessh"
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
}```
