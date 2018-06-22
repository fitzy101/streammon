package stream

import (
	"errors"
	"fmt"
	"os"
)

var (
	// numSubs tracks the number of subscribers. Will close when all are
	// closed.
	numSubs = 0
)

// RW implements the Subscriber and Publisher interfaces for a given
// Stream, allowing for communication between the interested parties.
type RW struct {
	stream   *Stream
	streamer chan string
	err      error
}

// NewSubscriber returns an encapsulated RW allowing the consumer to
// subscribe to text coming from the Stream.
func NewSubscriber(s *Stream) Subscriber {
	srw := RW{
		stream:   s,
		streamer: s.lines,
	}
	numSubs++
	return &srw
}

// NewPublisher returns an encapsulated RW allowing the consumer to
// subscribe to text coming from the Stream.
func NewPublisher(s *Stream) Publisher {
	srw := RW{
		stream:   s,
		streamer: s.lines,
	}
	return &srw
}

// Subscribe returns a channel where text will be sent unless closed.
func (srw *RW) Subscribe() chan string {
	go func() {
		srw.stream.readLines()
	}()
	return srw.streamer
}

// Publish sends a string to the channel that Subscribers will recieve.
func (srw *RW) Publish(line string) {
	srw.streamer <- line
	return
}

// Err returns any errors associated with the RW.
func (srw *RW) Err() error {
	return srw.err
}

// Close finishes the channel for any Subscribers.
func (srw *RW) Close() {
	close(srw.streamer)
	srw.err = errors.New("streamer closed for publishing")
	numSubs--
	if numSubs == 0 {
		fmt.Fprintln(os.Stdout, "No more files to watch, closing.")
		os.Exit(0)
	}
	return
}
