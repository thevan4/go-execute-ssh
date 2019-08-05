package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/logger"
	executessh "github.com/thevan4/go-execute-ssh/execute-ssh"
	"golang.org/x/crypto/ssh"
)

func main() {
	logging, err := logger.NewLogger("stdout", "info", "default", "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	host := "1.1.1.1:22"
	user := "admin4eg"
	password := "pass"
	shellPrompt := "#" //or $ for example
	timeoutForExecuteCommand := time.Duration(2 * time.Second)
	commands := []string{"show hostname", "show interface mgmt 0 | json"}
	var maxBufferBytes uint = 1000

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
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string,
			remote net.Addr,
			key ssh.PublicKey) error {
			return nil
		}),
	}

	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, err
	}
	return sshClient, nil
}
