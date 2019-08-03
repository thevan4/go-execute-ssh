# go-execute-ssh

[![Go Report Card](https://goreportcard.com/badge/github.com/thevan4/go-execute-ssh)](https://goreportcard.com/report/github.com/thevan4/go-execute-ssh) [![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
Executing commands on remote hosts via ssh.

To start, you must specify:

1. Shell prompt - **string**

2. Timeout for execute command - **time.Duration**

3. Commands - **[]string**

Execute result: **map[string]string** (key - command, value - result) and **error**

## Example

```golang
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
}```
