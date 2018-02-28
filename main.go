// main handles the command line validation and configuration of the underlying
// stream to watch. The contains the main run loop in watchStream().
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	re "regexp"
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
	log       bool
	// config    string
)

const (
	dfilepath  = "a full path to a file to monitor."
	ddelimeter = "a delimeter to split a matching line."
	dregexp    = "a regular expression to match."
	dcommand   = "a command to run after a match is found."
	dargs      = "a quoted string of arguments to the command."
	dlog       = "an option to turn on log output"
	// dtimeout   = "a timeout to wait before running the command."
	// dconfig    = "a configuration file to read from."
)

func usage() string {
	var sbuff bytes.Buffer
	sbuff.WriteString("Usage: streammon [OPTIONS]...\n")
	sbuff.WriteString(fmt.Sprintf("\t\t-f/--file: %s\n", dfilepath))
	sbuff.WriteString(fmt.Sprintf("\t\t-d/--delimeter %s\n", ddelimeter))
	sbuff.WriteString(fmt.Sprintf("\t\t-r/--regexp %s\n", dregexp))
	sbuff.WriteString(fmt.Sprintf("\t\t-c/--command %s\n", dcommand))
	sbuff.WriteString(fmt.Sprintf("\t\t-a/--args %s\n", dargs))
	sbuff.WriteString(fmt.Sprintf("\t\t-l %s\n", dlog))
	return sbuff.String()
}

// init: setup the command line flags & usage message.
func init() {

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

	// --args, -a
	flag.StringVar(&cargs, "args", "", dargs)
	flag.StringVar(&cargs, "a", "", dargs)

	// -l
	flag.BoolVar(&log, "l", false, dlog)

	// TODO: Implement.
	// --timeout, -t
	//flag.IntVar(&timeout, "timeout", 0, dtimeout)
	//flag.IntVar(&timeout, "t", 0, dtimeout)

	// --config, -cfg
	//flag.StringVar(&config, "config", "$HOME/.streammon.conf", dconfig)
	//flag.StringVar(&config, "cfg", "$HOME/.streammon.conf", dconfig)
}

// streamArgs holds the user provided arguments for validation.
type streamArgs struct {
	filepath  string
	delimeter string
	regexp    string
	command   string
	args      []string
	// timeout   int
	// config    string
}

func constructArgs() streamArgs {
	a := streamArgs{
		filepath:  filepath,
		delimeter: delimeter,
		regexp:    regexp,
		command:   command,
		// timeout:   timeout,
		// config:    config,
	}

	// Split up the arguments, we're expecting a quoted string.
	// Drop any of the other trailing flags.
	f := strings.Split(cargs, " ")
	for _, arg := range f {
		a.args = append(a.args, arg)
	}
	if err := validate(&a); err != nil {
		exitErr(err.Error())
	}
	return a
}

var (
	errFilepath = "a file must be provided or piped through stdin."
	errRegexp   = "you must provide a valid regular expression"
	errCommand  = "you must provide a command to run"
)

func validate(a *streamArgs) error {
	if a.filepath == "" && !isStdin() {
		return errors.New(errFilepath)
	}
	if a.regexp == "" {
		return errors.New(errRegexp)
	}
	// Check if its valid regexp
	_, err := re.Compile(a.regexp)
	if err != nil {
		return errors.New(errRegexp)
	}

	if a.command == "" {
		return errors.New(errCommand)
	}
	return nil
}

// isStdin returns true when file has data piped from stdin.
func isStdin() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func main() {
	flag.Usage = func() {
		exitErr(usage())
	}
	flag.Parse()

	if len(os.Args) == 1 {
		exitErr(usage())
	}

	strArgs := constructArgs()
	s, err := stream.NewStream(
		strArgs.regexp,
		strArgs.command,
		strArgs.delimeter,
		strArgs.filepath,
		strArgs.args,
		// strArgs.timeout, TODO: implement.
	)
	if log {
		stream.LogDebug = true
	}
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
	srw := stream.NewReader(s)
	// Listen for the lines received.
	for line := range srw.Subscribe() {
		match := s.Regexp.MatchString(line)
		if match {
			if err := s.ExecStreamComm(line); err != nil {
				fmt.Fprintf(os.Stderr, "error exec command %s: \n", err.Error())
			}
		}
	}
}

func exitErr(err string) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
