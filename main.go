package main

import (
	"fmt"
	"os"

	"github.com/fitzy101/streammon/stream"
)

func main() {
	s, err := stream.NewStream(
		`[1].*`,
		"touch",
		" ",
		"/users/WORK/Desktop/fakelog",
		0,
		[]string{"/users/WORK/Desktop/testing/${2}"},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating stream.\n")
	} else {
		watchStream(s)
	}
}

// watchStream sets up a watch on the Stream provided, and matches lines against
// the Stream's regexp.
func watchStream(s *stream.Stream) {
	srw := stream.NewStreamReader(s)

	// Listen for the lines received.
	for line := range srw.Subscribe() {
		match := s.Regexp.MatchString(line)
		if match {
			// Before running the command, we need to replace field
			// tokens with the actual matched line fields.
			preppedArgs := s.PrepArgs(line)
			if err := s.ExecStreamComm(line, preppedArgs); err != nil {
				fmt.Fprintln(os.Stderr, "error exec command %s: ", err)
			}
		}
	}
}
