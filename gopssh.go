package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
)

// main program class
type GopSSH struct {
	cmd        string
	hosts      []string
	user       string
	key        string
	parallel   int
	timeout    int
	outdir     string
	errdir     string
	inline     bool
	inline_out bool
	send_input bool
	results    map[string]chan *CmdResult
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
	gopssh.cmd = ARGS.Command
	if ARGS.Hosts != "" {
		// add hosts from host file, if specified
		gopssh.hosts = append(gopssh.hosts, *read_hosts_from_host_file()...)
	}
	for _, e := range ARGS.Host {
		if !slices.Contains(gopssh.hosts, e) {
			gopssh.hosts = append(gopssh.hosts, e)
		}
	}
	if ARGS.User == "" {
		gopssh.user = os.Getenv("USER")
	}
	gopssh.key = ARGS.Key
	if ARGS.Par == 0 {
		gopssh.parallel = runtime.NumCPU()
	}
	gopssh.timeout = ARGS.Timeout
	gopssh.outdir = ARGS.Outdir
	gopssh.errdir = ARGS.Errdir
	gopssh.inline = ARGS.Inline
	gopssh.inline_out = ARGS.Inline_Stdout
	gopssh.send_input = ARGS.Send_Input
}

func (gopssh *GopSSH) exec_command() {
	for _, host := range gopssh.hosts {
		ssh_client := SSHClient{
			hostname: host,
			port:     22,
			user:     gopssh.user,
			key:      gopssh.key,
		}
		ch := make(chan *CmdResult)
		gopssh.results[host] = ch
		go ssh_client.ExecCmd(gopssh.cmd, ch)
	}
}

// take results from channels and print out
// returns exit code 0 if all went fine, 0xff otherwise
func (gopssh *GopSSH) print_out_results() int {
	exit_code := 0
	for _, host := range gopssh.hosts {
		result := <-gopssh.results[host]
		fmt.Printf("[%s] STDOUT>\n%s\n", host, result.stdout)
		fmt.Printf("[%s] STDERR>\n%s\n", host, result.stderr)
	}
	return exit_code
}

/* functions */

// read host file (if any) and return list of hostnames
func read_hosts_from_host_file() *[]string {
	content, err := os.ReadFile(ARGS.Hosts)
	if err != nil {
		log.Panicf("unable to read host file: %s", ARGS.Hosts)
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
	return &result
}
