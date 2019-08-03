package executessh

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Connection ...
type Connection struct {
	*ssh.Client
	password string
}

// Connect ...
func Connect(addr, user, password string) (*Connection, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }), // TODO: make customization possible
	}

	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, err
	}

	return &Connection{conn, password}, nil

}

// Еhis hardcode should be. 500000 - is the width and height of the pseudo terminal
// In general, this is not important, since we only read real data
const (
	pseudoTerminalWidth  = 500000
	pseudoTerminalHeight = 500000
)

// SendCommands ...
func (conn *Connection) SendCommands(shellPrompt string, timeoutSeconds time.Duration, maxBufferBytes uint, commands ...string) ([]string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	sshOut, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to setup stdin for session: %v", err)
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to setup stdout for session: %v", err)
	}

	if err = session.RequestPty("xterm", pseudoTerminalHeight, pseudoTerminalWidth, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("request for pseudo terminal failed: %s", err)
	}
	if err = session.Shell(); err != nil {
		session.Close()
		return nil, fmt.Errorf("request for shell failed: %v", err)
	}

	_, err = readExpectedBuff(shellPrompt, "", sshOut, timeoutSeconds, maxBufferBytes) // reset everything to start shellPrompt
	if err != nil {
		return nil, err
	}

	results := []string{}
	for _, command := range commands {
		err = writeBuff(command, sshIn) // run command. simple sending buffer to sshIn
		if err != nil {
			return nil, fmt.Errorf("failed to run: %s", err)
		}

		currentResult, err := readExpectedBuff("\r", command+"\r"+"\r"+"\n", sshOut, timeoutSeconds, maxBufferBytes)
		if err != nil {
			return nil, fmt.Errorf("can't read expected buffer `\r`: %v", err)
		}

		results = append(results, strings.TrimSpace(currentResult))
		_, err = readExpectedBuff(shellPrompt, "", sshOut, timeoutSeconds, maxBufferBytes) // reset everything to start shellPrompt
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func readExpectedBuff(whatDoExpect, whatToSkip string, sshOut io.Reader, timeoutSeconds time.Duration, maxBufferBytes uint) (string, error) {
	resultChan := make(chan string, 1)
	defer close(resultChan)
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds)
	defer cancel()
	errorChan := make(chan error, 1)
	defer close(errorChan)

	go readBuffForExpectedString(whatDoExpect, whatToSkip, sshOut, resultChan, errorChan, maxBufferBytes)

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("waiting for execute command took longer than %v seconds", timeoutSeconds)
	}
}

func readBuffForExpectedString(whatDoExpect, whatToSkip string, sshOut io.Reader, resultChan chan<- string, errorChan chan error, maxBufferBytes uint) {
	var waitingString string
	var n int
	var err error
	var checkWhatToSkip = true
	buf := make([]byte, maxBufferBytes)

takeBuffer:
	for !strings.Contains(waitingString, whatDoExpect) {
		n, err = sshOut.Read(buf) //this reads the ssh terminal
		if err != nil {
			errorChan <- err
			return
		}
		waitingString += string(buf[:n])
	}

	if checkWhatToSkip { // if run command already droped - do not compare string again
		if waitingString == whatToSkip { // if read console equal run command - skip it
			waitingString = ""
			checkWhatToSkip = false
			goto takeBuffer
		}
	}

	resultChan <- waitingString
}

func writeBuff(command string, sshIn io.WriteCloser) error {
	_, err := sshIn.Write([]byte(command + "\r"))
	return err
}
