// main handles the command line validation and configuration of the underlying
// stream to watch. The contains the main run loop in watchStream().
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	re "regexp"
	"strings"
	"sync"

	"github.com/fitzy101/streammon/internal/stream"
)

var (
	filepath  string
	delimiter string
	regexp    string
	command   string
	cargs     string
	log       bool
	config    string
)

const (
	dfilepath  = "a full path to a file to monitor."
	ddelimiter = "a delimiter to split a matching line."
	dregexp    = "a regular expression to match."
	dcommand   = "a command to run after a match is found."
	dargs      = "a quoted string of arguments to the command."
	dlog       = "an option to turn on log output"
	dconfig    = "a configuration file to read from, all other flags are ignored."
	// dtimeout   = "a timeout to wait before running the command."
)

func usage() string {
	var sbuff bytes.Buffer
	sbuff.WriteString("Usage: streammon [OPTIONS]...\n")
	sbuff.WriteString(fmt.Sprintf("\t\t-f/--file: %s\n", dfilepath))
	sbuff.WriteString(fmt.Sprintf("\t\t-d/--delimiter %s\n", ddelimiter))
	sbuff.WriteString(fmt.Sprintf("\t\t-r/--regexp %s\n", dregexp))
	sbuff.WriteString(fmt.Sprintf("\t\t-c/--command %s\n", dcommand))
	sbuff.WriteString(fmt.Sprintf("\t\t-a/--args %s\n", dargs))
	sbuff.WriteString(fmt.Sprintf("\t\t-k/--config %s\n", dconfig))
	sbuff.WriteString(fmt.Sprintf("\t\t-l %s\n", dlog))
	return sbuff.String()
}

// init: setup the command line flags & usage message.
func init() {
	// --file, -f
	flag.StringVar(&filepath, "file", "", dfilepath)
	flag.StringVar(&filepath, "f", "", dfilepath)

	// --delimiter, -d
	flag.StringVar(&delimiter, "delimiter", " ", ddelimiter)
	flag.StringVar(&delimiter, "d", " ", ddelimiter)

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

	// --config, -k
	flag.StringVar(&config, "config", "", dconfig)
	flag.StringVar(&config, "k", "", dconfig)

	// TODO: Implement.  --timeout, -t
	//flag.IntVar(&timeout, "timeout", 0, dtimeout)
	//flag.IntVar(&timeout, "t", 0, dtimeout)
}

// streamArgs holds the user provided arguments for validation.
type streamArgs struct {
	filepath  string
	delimiter string
	regexp    string
	command   string
	args      []string

	// timeout   int
}

// readFromFile retrieves the contents from fileP and returns the []byte.
func readFromFile(fileP string) ([]byte, error) {
	resp := []byte{}

	file, err := os.Open(fileP)
	defer file.Close()
	if err != nil {
		return resp, err
	}

	resp, err = ioutil.ReadAll(file)
	return resp, err
}

// parseConfigFile reads the config from input, and validates each of the
// stream's configuration.
func parseConfigFile(cFile []byte) ([]streamArgs, error) {
	resp := []streamArgs{}

	if len(cFile) == 0 {
		return resp, errors.New(errConfig)
	}

	type cfgArgs struct {
		Filepath  string `json:"filepath"`
		Delimiter string `json:"delimiter"`
		Regexp    string `json:"regexp"`
		Command   string `json:"command"`
		Args      string `json:"args"`
	}
	allConf := make([]cfgArgs, 0)

	if err := json.Unmarshal(cFile, &allConf); err != nil {
		return resp, errors.New(errConfig)
	}

	for _, c := range allConf {
		arg, err := constructArgs(
			c.Filepath,
			c.Delimiter,
			c.Regexp,
			c.Command,
			c.Args)
		if err != nil {
			return resp, errors.New(errConfigInvalid)
		}

		resp = append(resp, arg)
	}

	return resp, nil
}

// constructArgs validates the command line arguments and returns a valid
// streamArgs for making a stream.
func constructArgs(fp, dl, re, cmd, args string) (streamArgs, error) {

	a := streamArgs{
		filepath:  fp,
		delimiter: dl,
		regexp:    re,
		command:   cmd,
	}

	// Parse the provided arguments from left to right. The argument is either
	// space separated strings or a quoted string with spaces preserved.
	parser := func(r io.Reader) []string {
		var buf bytes.Buffer
		quote := '\''
		space := ' '
		eof := rune(0)
		ret := []string{}
		argsRd := bufio.NewReader(r)

		// Wraps any ReadRune() error with eof to stop parsing.
		read := func(a *bufio.Reader) rune {
			ch, _, err := argsRd.ReadRune()
			if err != nil {
				return eof
			}
			return ch
		}

		for {
			if ch := read(argsRd); ch == eof {
				break
			} else if ch == space {
				continue
			} else if ch == quote {
				buf.WriteRune(ch)
				for ch = read(argsRd); ch != quote && ch != eof; ch = read(argsRd) {
					buf.WriteRune(ch)
				}
				buf.WriteRune(quote) // preserve the last quote
			} else {
				buf.WriteRune(ch)
				for ch = read(argsRd); ch != space && ch != eof; ch = read(argsRd) {
					buf.WriteRune(ch)
				}
			}
			ret = append(ret, buf.String())
			buf.Reset()
		}

		return ret
	}

	a.args = parser(strings.NewReader(args))

	if err := validate(&a); err != nil {
		return a, err
	}

	return a, nil
}

var (
	errFilepath      = "a file must be provided or piped through stdin"
	errRegexp        = "you must provide a valid regular expression"
	errCommand       = "you must provide a command to run"
	errConfig        = "the config file was empty or contained invalid json"
	errConfigInvalid = "the config file contained invalid streammon config"
)

func validate(a *streamArgs) error {
	// We need a filepath if we're watching streams specified in a config
	// file.
	if isCfgFile(config) {
		if a.filepath == "" {
			return errors.New(errFilepath)
		}
	} else {
		if a.filepath == "" && !isStdin() {
			return errors.New(errFilepath)
		}
	}

	// Not much point without a regexp to look for.
	if a.regexp == "" {
		return errors.New(errRegexp)
	}

	// Check if its valid regexp
	_, err := re.Compile(a.regexp)
	if err != nil {
		return errors.New(errRegexp)
	}

	// We're the same as 'tail', without a command.
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

// isCfgFile return true when a configuration file has been specified. Used
// to distinguish between a Stdin read and a filepath read.
func isCfgFile(cfg string) bool {
	if cfg == "" {
		return false
	}
	return true
}

// getStreams constructs the streams based on configuration and returns the
// stream.Stream array as required.
func getStreams(cfg, filepath, delimiter, regexp, command, cargs string) ([]*stream.Stream, error) {
	var streams []*stream.Stream

	// If there is a config file, ignore other flags and validate the config
	// file options.
	if isCfgFile(cfg) {
		contents, err := readFromFile(cfg)
		if err != nil {
			return streams, err
		}

		strs, err := parseConfigFile(contents)
		if err != nil {
			return streams, err
		}

		// Make streams for all of the configured files.
		for _, str := range strs {
			s, err := stream.NewStream(
				str.regexp,
				str.command,
				str.delimiter,
				str.filepath,
				str.args,
			)
			streams = append(streams, s)
			if err != nil {
				return streams, err
			}
		}
	} else {
		strArgs, err := constructArgs(filepath, delimiter, regexp, command, cargs)
		if err != nil {
			exitErr(err.Error())
		}

		s, err := stream.NewStream(
			strArgs.regexp,
			strArgs.command,
			strArgs.delimiter,
			strArgs.filepath,
			strArgs.args,
		)
		if err != nil {
			return streams, err
		}
		streams = append(streams, s)
	}
	return streams, nil
}

func main() {
	flag.Usage = func() {
		exitErr(usage())
	}
	flag.Parse()

	if len(os.Args) == 1 {
		exitErr(usage())
	}

	streams, err := getStreams(config, filepath, delimiter, regexp, command, cargs)
	if err != nil {
		exitErr(err.Error())
	}

	if log {
		stream.LogDebug = true
	}

	var wg sync.WaitGroup
	for _, s := range streams {
		wg.Add(1)
		go watchStream(s, &wg)
	}

	wg.Wait()
	os.Exit(0)
}

// watchStream sets up a watch on the Stream provided, and matches lines against
// the Stream's regexp.
func watchStream(s *stream.Stream, wg *sync.WaitGroup) {
	srw := stream.NewSubscriber(s)

	// Listen for the lines received.
	for line := range srw.Subscribe() {
		match := s.Regexp.MatchString(line)
		if match {
			if err := s.ExecStreamComm(line); err != nil {
				fmt.Fprintf(os.Stderr, "error exec command %s: \n", err.Error())
			}
		}
	}
	wg.Done()
}

func exitErr(err string) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
