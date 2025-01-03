package main

import (
	"bytes"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

/* Global vars */
var SIGNER *ssh.Signer

/*-------------------*/

type CmdResult struct {
	stdout    string
	stderr    string
	exit_code int
	duration  float32
}

type SSHClient struct {
	hostname string
	port     int
	user     string
	key      string
	ref      *ssh.Client
	session  *ssh.Session
}

func get_signer(key_path *string) (*ssh.Signer, error) {
	if SIGNER != nil {
		// return cached value
		return SIGNER, nil
	}
	key, err := os.ReadFile(*key_path)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}
	return &signer, nil
}

func (client *SSHClient) connect() error {
	var err error
	signer, _ := get_signer(&client.key)
	config := &ssh.ClientConfig{
		User: client.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(*signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client.ref, err = ssh.Dial("tcp", client.hostname+":22", config)
	return err
}

func (client *SSHClient) disconnect() {
	client.ref.Close()
}

func (client *SSHClient) run(cmd string) *CmdResult {
	var out, err bytes.Buffer
	client.session.Stderr = &err
	client.session.Stdout = &out
	if err := client.session.Run(cmd); err != nil {
		return &CmdResult{
			stdout: "",
			stderr: err.Error(),
		}
		//
	}
	return &CmdResult{
		stdout: out.String(),
		stderr: err.String(),
	}
}

func (client *SSHClient) ExecCmd(cmd string, ch chan *CmdResult) {
	var err error
	if err = client.connect(); err == nil {
		defer client.disconnect()
		if client.session, err = client.ref.NewSession(); err == nil {
			defer client.session.Close()
			ch <- client.run(cmd)
			return
		}
	}
	ch <- &CmdResult{"", err.Error(), 0x0f, 0}
}

// run ssh command on remote host
// and return ro channel for result
// func ssh_run(host string, cmd string) <-chan *CmdResult {
// map[CmdResult]string{}
// }
