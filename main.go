package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fitzy101/streammon/stream"
)

var (
	filepath  string
	delimeter string
	regexp    string
	command   string
	timeout   int
	cargs     string
	// config    string
)

// init: setup the command line flags & usage message.
func init() {
	const (
		dfilepath  = "a full path to a file to monitor."
		ddelimeter = "a delimeter to split a matchin line."
		dregexp    = "a regular expression to match."
		dcommand   = "a command to run after a match is found."
		dtimeout   = "a timeout to wait before running the command."
		dargs      = "a quoted string of arguments to the command."
		// dconfig    = "a configuration file to read from."
	)

	// --file, -f
	flag.StringVar(&filepath, "file", "", dfilepath)
	flag.StringVar(&filepath, "f", "", dfilepath)

	// --delimeter, -d
	flag.StringVar(&delimeter, "delimeter", " ", ddelimeter)
	flag.StringVar(&delimeter, "d", " ", ddelimeter)

	// --regexp, -r
	flag.StringVar(&regexp, "regexp", ".*", dregexp)
	flag.StringVar(&regexp, "r", ".*", dregexp)

	// --command, -c
	flag.StringVar(&command, "command", "", dcommand)
	flag.StringVar(&command, "c", "", dcommand)

	// --timeout, -t
	flag.IntVar(&timeout, "timeout", 0, dtimeout)
	flag.IntVar(&timeout, "t", 0, dtimeout)

	// --args, -a
	flag.StringVar(&cargs, "args", "", dargs)
	flag.StringVar(&cargs, "a", "", dargs)

	// --config, -cfg
	//flag.StringVar(&config, "config", "$HOME/.streammon.conf", "a configuration file to read from.")
	//flag.StringVar(&config, "cfg", "$HOME/.streammon.conf", "a configuration file to read from.")
}

// streamArgs holds the user provided arguments for validation.
type streamArgs struct {
	filepath  string
	delimeter string
	regexp    string
	command   string
	timeout   int
	args      []string
	// config    string
}

func constructArgs() streamArgs {
	a := streamArgs{
		filepath:  filepath,
		delimeter: delimeter,
		regexp:    regexp,
		command:   command,
		timeout:   timeout,
		//	config:    config,
	}

	// Split up the arguments, we're expecting a quoted string.
	// Drop any of the other trailing flags.
	f := strings.Split(cargs, " ")
	for _, arg := range f {
		a.args = append(a.args, arg)
	}
	if str, ok := validate(&a); !ok {
		exitErr(str)
	}
	return a
}

func validate(a *streamArgs) (string, bool) {
	if a.filepath == "" && !isStdin() {
		return "A file must be provided (or through a pipe).\n", false
	}
	if a.regexp == "" {
		return "You must provide a regular expression.\n", false
	}
	if a.command == "" {
		return "You must provide a command to run.\n", false
	}
	return "", true
}

// isStdin returns true when file has data piped from stdin.
func isStdin() bool {
	return true
	//	stat, _ := os.Stdin.Stat()
	//	return (stat.Mode() & os.ModeCharDevice) == 0
}

func main() {
	flag.Parse()
	strArgs := constructArgs()
	s, err := stream.NewStream(
		strArgs.regexp,
		strArgs.command,
		strArgs.delimeter,
		strArgs.filepath,
		strArgs.timeout,
		strArgs.args,
	)
	if err != nil {
		exitErr("error creating stream.\n")
	} else {
		watchStream(s)
	}
	os.Exit(0)
}

// watchStream sets up a watch on the Stream provided, and matches lines against
// the Stream's regexp.
func watchStream(s *stream.Stream) {
	srw := stream.NewStreamReader(s)
	// Listen for the lines received.
	for line := range srw.Subscribe() {
		match := s.Regexp.MatchString(line)
		if match {
			if err := s.ExecStreamComm(line); err != nil {
				fmt.Fprintln(os.Stderr, "error exec command %s: ", err)
			}
		}
	}
}

func exitErr(err string) {
	fmt.Fprintf(os.Stderr, err)
	os.Exit(1)
}
