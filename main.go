package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// global vars
var ARGS Args

//

type Args struct {
	Command  string
	Hostfile string
	Hosts    string
	User     string
	Key      string
	// Timeout       int
	// Outdir        string
	// Errdir        string
	// Inline        bool
	// Inline_Stdout bool
	Send_Input bool
}

func main() {
	// parser := arg.MustParse(&ARGS)
	// if len(os.Args) == 1 {
	// 	// no arguments were provided
	// 	parser.WriteHelp(os.Stderr)
	// 	return
	// }
	if err := parse_args(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := verify_args(&ARGS); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	gopssh := GopSSH{results: map[string]chan *CmdResult{}}
	os.Exit(gopssh.Main())
}

func parse_args() error {
	flags := flag.NewFlagSet("gopssh", flag.ExitOnError)
	flags.StringVar(&ARGS.Hostfile, "host_file", "", "read hosts from the given host_file")
	flags.StringVar(&ARGS.Hosts, "hosts", "", "add the given host strings to the list of hosts")
	flags.StringVar(&ARGS.User, "user", "", "use the given username")
	flags.StringVar(&ARGS.Key, "key", "", "SSH private key to authenticate with")
	err := flags.Parse(os.Args[1:])
	ARGS.Command = strings.Join(flags.Args(), " ")
	return err
}

func verify_args(args *Args) error {
	switch {
	case args.Command == "" && !args.Send_Input:
		return &EInvalidArgs{"please provide command or -I"}
	case args.Command != "" && args.Send_Input:
		return &EInvalidArgs{"please provide either command or -I, not both"}
	case args.Hostfile == "" && len(args.Hosts) == 0:
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
