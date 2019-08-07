# go-execute-ssh [![Go Report Card](https://goreportcard.com/badge/github.com/thevan4/go-execute-ssh)](https://goreportcard.com/report/github.com/thevan4/go-execute-ssh) [![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT) [![CircleCI](https://circleci.com/gh/thevan4/go-execute-ssh.svg?style=svg)](https://circleci.com/gh/thevan4/go-execute-ssh)

Executing commands on remote hosts via ssh.

To start, you must specify:

1. Shell prompt - **string**

2. Timeout for execute command - **time.Duration**

3. Commands - **[]string**

Execute result is struct:

```golang
type CommandAndResult struct {
    Command, Result string
}
```

And **error**

## Why execute result not map or array

Not a map, as in some cases the order of execution is important.

Not an array, since in some cases we don’t need all the results (sometimes the sequence of commands is important, but not their output).

## Usage

**WARNING! Demo run insecure!**
Build binary and run it with flags.

Short example:

```bash
go run main -host 1.1.1.1:22 -user admin1 -password pass -shellPrompt '$' -execute-timeout 10s -commands 'ls,cat 123.txt' -max-buffer-bytes 2000
```

Full example:

```bash
go run main -log-output syslog -log-level info -log-format default -syslog-tag sometag -host 1.1.1.1:22 -user admin1 -password pass -shellPrompt '$' -execute-timeout 10s -commands 'ls,cat 123.txt' -max-buffer-bytes 2000
```

All flags are not required; missing flags will be taken by default.

## Example for run

```golang
package main


import (
    "flag"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/sirupsen/logrus"
    "github.com/thevan4/go-billet/logger"
    executessh "github.com/thevan4/go-execute-ssh/execute-ssh"
    "golang.org/x/crypto/ssh"
)

var logOutput, logLevel, logFormat, syslogTag string
var host, user, password, shellPrompt, rawTimeoutForExecuteCommand, rawCommands, rawMaxBufferBytes string

func init() {
    flag.StringVar(&logOutput, "log-output", "stdout", "log output. Example: stdout")
    flag.StringVar(&logLevel, "log-level", "info", "log level. Example: info")
    flag.StringVar(&logFormat, "log-format", "default", "log format. Example: default")
    flag.StringVar(&syslogTag, "syslog-tag", "", "syslog tag. Example: sometag.")

    flag.StringVar(&host, "host", "127.0.0.1:22", "host and port for connect. Example: 127.0.0.1:22")
    flag.StringVar(&user, "user", "admin4eg", "username for connect. Example: admin4eg")
    flag.StringVar(&password, "password", "pass", "password for connect. Example: pass")
    flag.StringVar(&shellPrompt, "shell-prompt", "#", "shell prompt for remote terminal. Example: '#'")
    flag.StringVar(&rawTimeoutForExecuteCommand, "execute-timeout", "5s", "execute timeout. Example: 5s")
    flag.StringVar(&rawCommands, "commands", "show hostname,show interface mgmt 0 | json", "commands for execute. Delimiter ','. Example: show hostname,show interface mgmt 0 | json")
    flag.StringVar(&rawMaxBufferBytes, "max-buffer-bytes", "1000", "max buffer bytes for read execute result. Example: 1000")
    flag.Parse()
}

func main() {
    logging, err := logger.NewLogger(logOutput, logLevel, logFormat, syslogTag)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    timeoutForExecuteCommand, err := time.ParseDuration(rawTimeoutForExecuteCommand)
    if err != nil {
        logging.Fatalf("fail to parse timeout for execute command: %v", err)
    }

    commands := strings.Split(rawCommands, ",")

    u64, err := strconv.ParseUint(rawMaxBufferBytes, 10, 32)
    if err != nil {
        logging.Fatalf("fail to parse max-buffer-bytes: %v", err)
    }
    maxBufferBytes := uint(u64)

    logging.WithFields(logrus.Fields{
        "Log output":                  logOutput,
        "Log level":                   logLevel,
        "Log format":                  logFormat,
        "Log:syslog tag":              syslogTag,
        "Host":                        host,
        "User":                        user,
        "Password":                    password,
        "Shell prompt":                shellPrompt,
        "Timeout for execute command": timeoutForExecuteCommand,
        "Commands":                    commands,
        "Max buffer bytes":            maxBufferBytes,
    }).Info("Start parameters is:")

    sshClient, err := newSSHClient(host, user, password)
    if err != nil {
        logging.Fatal(err)
    }

    result, err := executessh.SendCommands(sshClient, shellPrompt, timeoutForExecuteCommand, maxBufferBytes, commands...)
    if err != nil {
        logging.Fatal(err)
    }

    for num, commandAndResult := range result {
        logging.WithFields(logrus.Fields{
            "command": commandAndResult.Command,
            "result":  commandAndResult.Result,
        }).Infof("execute number %v", num+1)
    }
}

func newSSHClient(addr, user, password string) (*ssh.Client, error) {
    sshConfig := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    sshClient, err := ssh.Dial("tcp", addr, sshConfig)
    if err != nil {
        return nil, err
    }
    return sshClient, nil
}
```
