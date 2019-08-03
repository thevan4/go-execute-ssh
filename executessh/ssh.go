package executessh

import (
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
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
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
func (conn *Connection) SendCommands(shellPrompt string, timeoutSeconds int, commands ...string) ([]string, error) {
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

	_, err = readExpectedBuff(shellPrompt, "", sshOut, 2) // reset everything to start shellPrompt
	if err != nil {
		return nil, err
	}

	results := []string{}
	for _, command := range commands {
		err = writeBuff(command, sshIn) // run command. simple sending buffer to sshIn
		if err != nil {
			return nil, fmt.Errorf("failed to run: %s", err)
		}

		oneResult, err := readExpectedBuff("\r", command+"\r"+"\r"+"\n", sshOut, 2)
		if err != nil {
			return nil, fmt.Errorf("can't read expected buffer `\r`: %v", err)
		}

		results = append(results, strings.TrimSpace(oneResult))
		_, err = readExpectedBuff(shellPrompt, "", sshOut, 2) // reset everything to start shellPrompt
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func readExpectedBuff(whattoexpect, whattoskip string, sshOut io.Reader, timeoutSeconds int) (string, error) {
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	defer close(ch)
	defer close(errCh)
	go func(whattoexpect string, sshOut io.Reader) {
		buffRead := make(chan string)
		go readBuffForExpectedString(whattoexpect, whattoskip, sshOut, buffRead)
		select {
		case ret := <-buffRead:
			ch <- ret
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			errCh <- fmt.Errorf("waiting for execute command took longer than %v seconds", timeoutSeconds)
		}
	}(whattoexpect, sshOut)

	select {
	case result := <-ch:
		return result, nil
	case err := <-errCh:
		return "", err
	}
}

func readBuffForExpectedString(whattoexpect, whattoskip string, sshOut io.Reader, buffRead chan<- string) {
	buf := make([]byte, 1000)
	n, err := sshOut.Read(buf) //this reads the ssh terminal
	waitingString := ""
	if err == nil {
		waitingString = string(buf[:n])
	}

takeBuffer:
	for (err == nil) && (!strings.Contains(waitingString, whattoexpect)) {
		n, err = sshOut.Read(buf)
		waitingString += string(buf[:n])
		//fmt.Println(waitingString) // for debug

	}
	if waitingString == whattoskip { // if read console equal run command - skip it
		waitingString = ""
		goto takeBuffer
	}

	buffRead <- waitingString
}

func writeBuff(command string, sshIn io.WriteCloser) error {
	_, err := sshIn.Write([]byte(command + "\r"))
	return err
}
