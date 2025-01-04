package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

// main program class
type GopSSH struct {
	Args
	results map[string]chan *CmdResult
}

// main
func (gopssh *GopSSH) Main() int {
	gopssh.read_args()
	gopssh.exec_command()
	return gopssh.print_out_results()
}

/* other methods */

// populate GopSSH fields from cmdline args
func (gopssh *GopSSH) read_args() {
	gopssh.Command = ARGS.Command
	if ARGS.Hostfile != "" {
		// add hosts from host file, if specified
		gopssh.Hosts += read_hosts_from_host_file()
	}
	for _, e := range strings.Fields(ARGS.Hosts) {
		if !slices.Contains(strings.Fields(gopssh.Hosts), e) {
			gopssh.Hosts += e + " "
		}
	}
	if ARGS.User == "" {
		gopssh.User = os.Getenv("USER")
	}
	gopssh.Key = ARGS.Key
	// gopssh.Timeout = ARGS.Timeout
	// gopssh.Outdir = ARGS.Outdir
	// gopssh.Errdir = ARGS.Errdir
	// gopssh.Inline = ARGS.Inline
	// gopssh.Inline_Stdout = ARGS.Inline_Stdout
	gopssh.Send_Input = ARGS.Send_Input
}

func (gopssh *GopSSH) exec_command() {
	for _, host := range strings.Fields(gopssh.Hosts) {
		ssh_client := SSHClient{
			hostname: host,
			port:     22,
			user:     gopssh.User,
			key:      gopssh.Key,
		}
		ch := make(chan *CmdResult)
		gopssh.results[host] = ch
		go ssh_client.ExecCmd(gopssh.Command, ch)
	}
}

// take results from channels and print out
// returns exit code 0 if all went fine, 0xff otherwise
func (gopssh *GopSSH) print_out_results() int {
	exit_code := 0
	var result *CmdResult
	for _, host := range strings.Fields(gopssh.Hosts) {
		result = <-gopssh.results[host]
		fmt.Printf("[%s] STDOUT:\n%s\n", host, result.stdout)
		fmt.Printf("[%s] STDERR:\n%s\n", host, result.stderr)
	}
	return exit_code
}

/* functions */

// read host file (if any) and return list of hostnames
func read_hosts_from_host_file() string {
	content, err := os.ReadFile(ARGS.Hostfile)
	if err != nil {
		log.Panicf("unable to read host file: %s", ARGS.Hostfile)
	}
	// replacing spaces with newlines
	for i, b := range content {
		if b == ' ' {
			content[i] = '\n'
		}
	}
	lines := bytes.Split(content, []byte("\n"))
	result := []string{}
	for _, line := range lines {
		// removing trailing spaces
		trimmed := string(bytes.TrimSpace(line))
		// ignoring newline in EOF
		if len(trimmed) != 0 && !slices.Contains(result, trimmed) {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, " ")
}
