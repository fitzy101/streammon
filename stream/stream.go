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

// Stream holds the information for the monitored stream.
type Stream struct {
	File    string
	Cmd     string
	Args    []string
	Regexp  *regexp.Regexp
	Timeout int
	Delim   string
	Fields  []int
	Lines   chan string
}

type StreamReader interface{}

var (
	FieldDelim = `\$\{[0-9]*\}`
)

func NewStream(pattern, cmd, delim, file string, timeout int, args []string) (*Stream, error) {
	s := Stream{
		Cmd:     "touch",
		Args:    []string{"/users/WORK/Desktop/${0}"},
		Delim:   " ",
		File:    file,
		Timeout: timeout,
	}
	reg, err := setupRegexp(pattern)
	if err != nil {
		return nil, err
	}
	s.Regexp = reg
	s.Fields = parseFields(s.Args)
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
func (s *Stream) tailFile() chan string {
	conf := tail.Config{
		Follow: true,
		ReOpen: true,
	}
	t, err := tail.TailFile(s.File, conf)

	// Resend the lines back to the listener.
	lines := make(chan string)
	go func() {
		for line := range t.Lines {
			lines <- line.Text
			if line.Err != nil {
				fmt.Fprintln(os.Stderr, "error reading %s: ", err)
			} else {

			}
		}
	}()
	return lines
}

// ReadLines creates a string channel that the lines of the file
// will be sent to.
func (s *Stream) ReadLines() chan string {
	lines := make(chan string)
	if s.File == "" {
		// We're reading from stdin.
		scanner := s.openScanner()
		go func() {
			for scanner.Scan() {
				lines <- scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "error reading %s: ", err)
			}
			close(lines)
		}()
		return lines
	} else {
		// We'll tail the file instead.
		return s.tailFile()
	}
}

// execStreamComm is called with a matched line from the Stream, and executes
// the command for that stream.
func (s *Stream) ExecStreamComm(matchLn string, args []string) error {
	cmd := exec.Command(s.Cmd, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	if out.String() != "" {
		fmt.Printf("output: %s matched line: %s.\n", out.String(), matchLn)
	}
	return nil
}

// splitMatch takes a line that matched the Stream's regexp, and splits it on
// the Streams delimiter. After that, it replaces any of the field tokens with
// the actual field.
func (s *Stream) PrepArgs(line string) []string {
	spl := strings.Split(line, s.Delim)
	preppedArgs := []string{}

	// For all of the arguments, we want to replace any of the field tokens
	// with the actual field. The output of this loop should be the arg
	// string with the log line including the actual field text instead of
	// the token.
	for _, arg := range s.Args {
		argStr := arg
		for _, field := range s.Fields {
			if len(spl) >= field {
				if field == 0 {
					argStr = insertField(argStr, line, field)
				} else {
					argStr = insertField(argStr, spl[field], field)
				}
			}
		}
		preppedArgs = append(preppedArgs, argStr)
	}
	return preppedArgs
}

// insertField replaces the field tokens with the field text.
// For example if the string was "this is my ${5} field" and the 5th field was
// "log", the output is "this is my log field".
func insertField(str, replace string, field int) string {
	fieldStr := fmt.Sprintf("${%v}", field)
	return strings.Replace(str, fieldStr, replace, -1)
}

// parseFields searches the arguments of a Stream for ${[0-9]} fields for
// the commands, and returns an []int of the fields.
func parseFields(args []string) []int {
	fields := []int{}
	for _, arg := range args {
		if match, _ := regexp.MatchString(FieldDelim, arg); match {
			// Are there any field tokens found?
			token := strings.Split(arg, `${`)[1]
			token = strings.Split(token, `}`)[0]
			if i, err := strconv.Atoi(token); err != nil {
				fmt.Fprintf(os.Stderr, "error parsing fields: %s", err)
			} else {
				fields = append(fields, i)
			}
		}
	}
	return fields
}

// setupRegexp compiles the regular expression included, and
// returns an error if it didn't compile.
func setupRegexp(pattern string) (*regexp.Regexp, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return r, err
	}
	return r, nil
}
