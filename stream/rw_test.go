package stream

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestSubscriber(t *testing.T) {
	s, err := NewStream(
		".*",
		"echo",
		":",
		"/dev/null",
		[]string{"foo", "bar"},
	)
	if err != nil {
		t.Errorf(err.Error())
	}

	srw := NewSubscriber(s)
	pub := NewPublisher(s)
	if srw.Err() != nil {
		t.Errorf(srw.Err().Error())
	}

	if getSubs() != 1 {
		t.Errorf("new subscribers in count when one exists")
	}
	addSub() // prevent the program from exiting when all subs close

	var wg sync.WaitGroup
	wg.Add(1)
	msgCount := 0
	expected := 10

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for line := range srw.Subscribe() {
			match := s.Regexp.MatchString(line)
			if match {
				if err := s.ExecStreamComm(line); err != nil {
					fmt.Fprintf(os.Stderr, "error exec command %s: \n", err.Error())
				}
			}
			msgCount++
		}
	}(&wg)

	// publish some messages to the srw
	ticker := time.NewTicker(5 * time.Millisecond)
	count := 0
outer:
	for {
		select {
		case <-ticker.C:
			if count == 10 {
				ticker.Stop()
				srw.Close()
				break outer
			}

			pub.Publish(fmt.Sprintf("up to count %v\n", count))
			count++
		}
	}
	wg.Wait()

	if msgCount != expected {
		t.Errorf("expected msgCount to be %v, got %v", expected, msgCount)
	}

	if err := srw.Err(); err.Error() != "streamer closed for publishing" {
		t.Errorf("expected stream closed error")
	}
}
