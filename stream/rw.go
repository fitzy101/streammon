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

// StreamRW implements the StreamReader and StreamWriter interfaces for a given
// Stream, allowing for communication between the interested parties.
type StreamRW struct {
	stream   *Stream
	streamer chan string
	err      error
}

// NewStreamReader returns an encapsulated StreamRW allowing the consumer to
// subscribe to text coming from the Stream.
func NewStreamReader(s *Stream) StreamReader {
	srw := StreamRW{
		stream:   s,
		streamer: s.lines,
	}
	numSubs++
	return &srw
}

// NewStreamWriter returns an encapsulated StreamRW allowing the consumer to
// subscribe to text coming from the Stream.
func NewStreamWriter(s *Stream) StreamWriter {
	srw := StreamRW{
		stream:   s,
		streamer: s.lines,
	}
	return &srw
}

// Subscribe returns a channel where text will be sent unless closed.
func (srw *StreamRW) Subscribe() chan string {
	go func() {
		srw.stream.readLines()
	}()
	return srw.streamer
}

// Publish sends a string to the channel that Subscribers will recieve.
func (srw *StreamRW) Publish(line string) {
	srw.streamer <- line
	return
}

// Err returns any errors associated with the StreamRW.
func (srw *StreamRW) Err() error {
	return srw.err
}

// Close finishes the channel for any Subscribers.
func (srw *StreamRW) Close() {
	close(srw.streamer)
	srw.err = errors.New("streamer closed for publishing")
	numSubs--
	if numSubs == 0 {
		fmt.Fprintln(os.Stdout, "No more files to watch, closing.")
		os.Exit(0)
	}
	return
}
