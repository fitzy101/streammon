// Package stream contains the main logic for handling actions within a stream,
// such as reading the lines retrieved, setting up the command to run, and
// executing the command.
package stream

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/hpcloud/tail"
)

var (
	// LogDebug controls the logging level, when true the stream will
	// write logs to stdout.
	LogDebug = false
)

// Stream holds the information for the monitored stream.
type Stream struct {
	Regexp *regexp.Regexp
	file   string
	cmd    string
	args   []string
	delim  string
	fields []int
	lines  chan string
	// timeout int
}

// Subscriber provides functions for a consumer of the Stream's output to
// subscribe, ie. receive text coming through the stream.
type Subscriber interface {
	Subscribe() chan string
	Err() error
	Close()
}

// Publisher provides functions to publish to any subscribers of a stream.
type Publisher interface {
	Publish(string)
	Err() error
	Close()
}

var (
	// FieldDelim - regex that matches a field pattern.
	FieldDelim = `.*\#\{[0-9]*\}.*`
)

// NewStream constructs a Stream for processing of a file. This calls the
// necessary field parsing functions before returning.
func NewStream(pattern, cmd, delim, file string, args []string) (*Stream, error) {
	s := Stream{
		cmd:   cmd,
		args:  args,
		delim: delim,
		file:  file,
		lines: make(chan string),
		// timeout: timeout,
	}
	reg, err := setupRegexp(pattern)
	if err != nil {
		return nil, err
	}
	s.fields = parseFields(s.args)
	s.Regexp = reg
	return &s, nil
}

// openScanner creates a new file scanner from the Stream -- we'll be reading from
// stdin as there was no file included.
func (s *Stream) openScanner() *bufio.Scanner {
	scanner := bufio.NewScanner(os.Stdin)
	return scanner
}

// openFile tails the Stream's file, returning the new lines back
// via string channel.
func (s *Stream) tailFile(swr Publisher) {
	conf := tail.Config{
		Follow: true,
		Poll:   true,
		Logger: tail.DiscardingLogger,
	}
	t, err := tail.TailFile(s.file, conf)

	// Catch any file closures that will cause a panic to unravel, so we
	// can close the subscribers nicely.
	go func() {
		// Resend the lines back to any Subscribers.
		for line := range t.Lines {
			if swr.Err() == nil {
				swr.Publish(line.Text)
			} else {
				swr.Close()
			}
			if line.Err != nil {
				fmt.Fprintf(os.Stderr, "error reading %s: ", err)
			}
		}
		// Close the channel, we're done tailing.
		swr.Close()
	}()
}

// readLines creates a string channel that the lines of the file
// will be sent to.
func (s *Stream) readLines() {
	swr := NewPublisher(s)
	if s.file == "" {
		// We're reading from stdin.
		scanner := s.openScanner()
		go func() {
			for scanner.Scan() {
				if swr.Err() == nil {
					swr.Publish(scanner.Text())
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "error reading %s: ", err)
			}
			// We've exhausted the scanner.
			swr.Close()
		}()
	} else {
		// Tail the file instead.
		s.tailFile(swr)
	}
}

// ExecStreamComm is called with a matched line from the Stream, and executes
// the command for that stream.
func (s *Stream) ExecStreamComm(matchLn string) error {
	// Before running the command, we need to replace field
	// tokens with the actual matched line fields.
	args := prepArgs(matchLn, s)

	if LogDebug {
		fmt.Printf("calling %s with args %v\n", s.cmd, args)
	}

	cmd := exec.Command(s.cmd, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}

	if out.String() != "" && LogDebug {
		fmt.Printf("output: %s matched line: %s.\n", out.String(), matchLn)
	}
	return nil
}

// prepArgs takes a line that matched the Stream's regexp, and splits it on
// the Streams delimiter. After that, it replaces any of the field tokens with
// the actual field.
func prepArgs(line string, s *Stream) []string {
	spl := strings.Split(line, s.delim)
	preppedArgs := []string{}

	// For all of the arguments, we want to replace any of the field tokens
	// with the actual field. The output of this loop should be the arg
	// string with the log line including the actual field text instead of
	// the token.
	for _, argStr := range s.args {
		for _, field := range s.fields {
			if len(spl) >= field {
				if field == 0 {
					argStr = insertField(argStr, line, field)
				} else {
					// Field tokens start from index 1 (unless
					// referring to the whole line).
					argStr = insertField(argStr, spl[field-1], field)
				}
			}
		}
		preppedArgs = append(preppedArgs, argStr)
	}
	return preppedArgs
}

// insertField replaces the field tokens with the field text.
// For example if the string was "this is my #{5} field" and the 5th field was
// "log", the output is "this is my log field".
func insertField(str, replace string, field int) string {
	fieldStr := fmt.Sprintf("#{%v}", field)
	return strings.Replace(str, fieldStr, replace, -1)
}

// parseFields searches the arguments of a Stream for #{[0-9]} fields for
// the commands, and returns an []int of the fields.
func parseFields(args []string) []int {
	fields := []int{}
	for _, arg := range args {
		ind := strings.Index(arg, `#{`)
		for ind != -1 {
			token := strings.Split(arg[ind+2:], `}`)[0]
			if i, err := strconv.Atoi(token); err != nil {
				fmt.Fprintf(os.Stderr, "error parsing fields: %s", err)
				break
			} else {
				fields = append(fields, i)
			}

			// There could be more than one, skip to the next token.
			arg = arg[ind+len(token)+3:]
			ind = strings.Index(arg, `#{`)
		}
	}
	return fields
}

// setupRegexp compiles the regular expression included, and returns an error
// the regex pattern didn't compile.
func setupRegexp(pattern string) (*regexp.Regexp, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return r, err
	}
	return r, nil
}
