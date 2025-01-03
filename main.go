package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

// global vars
var ARGS Args

//

type Args struct {
	Command       string   `arg:"positional" help:"command to run on remote hosts"`
	Hosts         string   `help:"read hosts from the given host_file" placeholder:"host_file"`
	Host          []string `arg:"-H" help:"add the given host strings to the list of hosts" placeholder:"host..."`
	User          string   `arg:"-l" help:"use the given username as the default for any host entries that don't specifically specify a user"`
	Key           string   `arg:"-k,required" help:"SSH private key"`
	Par           int      `arg:"-p" help:"use the given number as the maximum number of concurrent connections"`
	Timeout       int      `arg:"-t" help:"make connections time out after the given number of seconds (unless set to 0)"`
	Outdir        string   `arg:"-o" help:"save standard output to files in the given directory"`
	Errdir        string   `arg:"-e" help:"save standard error to files in the given directory"`
	Inline        bool     `arg:"-i" help:"display standard output and standard error as each host completes"`
	Inline_Stdout bool     `help:"display standard output (but not standard error) as each host completes"`
	Send_Input    bool     `arg:"-I" help:"read input and send to each ssh process"`
}

func main() {
	parser := arg.MustParse(&ARGS)
	if len(os.Args) == 1 {
		// no arguments were provided
		parser.WriteHelp(os.Stderr)
		return
	}
	if err := verify_args(&ARGS); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	gopssh := GopSSH{results: map[string]chan *CmdResult{}}
	os.Exit(gopssh.Main())
}

func verify_args(args *Args) error {
	switch {
	case args.Command == "" && !args.Send_Input:
		return &EInvalidArgs{"please provide command or -I"}
	case args.Command != "" && args.Send_Input:
		return &EInvalidArgs{"please provide either command or -I, not both"}
	case args.Hosts == "" && len(args.Host) == 0:
		return &EInvalidArgs{"please provide --host or/and --hosts"}
	default:
		return nil
	}
}

// errors

type EInvalidArgs struct{ msg string }

func (err *EInvalidArgs) Error() string {
	return "INVALID ARGS: " + err.msg
}
