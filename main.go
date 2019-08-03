package main

import (
	"fmt"
	"log"
	"time"

	"github.com/thevan4/go-execute-ssh/executessh"
)

func main() {
	host := "1.1.1.1:22"
	user := "admin4eg"
	password := "pass"
	shellPrompt := "#" //or $ for example
	timeoutSeconds := time.Duration(2 * time.Second)
	commands := []string{"show hostname", "show interface mgmt 0 | json"}
	var maxBufferBytes uint = 1000

	connection, err := executessh.Connect(host, user, password)
	if err != nil {
		log.Fatal(err)
	}

	output, err := connection.SendCommands(shellPrompt, timeoutSeconds, maxBufferBytes, commands...)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output)
}
