package main

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	testTable := []struct {
		args *streamArgs
		err  error
	}{
		{
			args: &streamArgs{},
			err:  errors.New(errFilepath),
		},
		{
			args: &streamArgs{
				filepath: "/home",
			},
			err: errors.New(errRegexp),
		},
		{
			args: &streamArgs{
				filepath: "/home",
				regexp:   ".*",
			},
			err: errors.New(errCommand),
		},
		{
			args: &streamArgs{
				filepath: "/home",
				regexp:   ".*",
				command:  "touch",
			},
		},
	}

	for _, table := range testTable {
		retErr := validate(table.args)
		if retErr == nil && table.err == nil {
			continue
		}
		if retErr == nil && table.err != nil {
			t.Errorf("No error returned, expected %v", table.err)
			continue
		}
		if retErr != nil && table.err == nil {
			t.Errorf("Error returned as %v, expected nil.", retErr)
			continue
		}
		if retErr.Error() != table.err.Error() {
			t.Errorf("Error type was incorrect, got %v, want %v.", retErr.Error(), table.err)
		}
	}

}

func TestConfig(t *testing.T) {
	testTable := []struct {
		config []byte
		err    error
	}{
		{
			config: []byte{},
			err:    errors.New(errConfig),
		},
		{
			config: []byte("Invalid JSON"),
			err:    errors.New(errConfig),
		},
		{
			config: []byte("[{}]"),
			err:    errors.New(errConfigInvalid),
		},
		{
			config: []byte(`[
				{
					"filepath":"/var/log/messages",
					"delimiter":" ",
					"regexp":"MATCHTHIS.*",
					"command":"redis-cli",
					"args":"publish key 'value'"
				}
			]`),
		},
		{
			config: []byte(`[
				{
					"filepath":"/var/log/messages",
					"delimiter":" ",
					"regexp":"MATCHTHIS.*",
					"command":"redis-cli",
					"args":"publish key 'value'"
				},
				{
					"filepath":"/var/log/nginx.log",
					"regexp":"GET.*\\ 200\\ ",
					"command":"redis-cli",
					"args":"publish key 'other value'"
				}
			]`),
		},
	}

	for _, table := range testTable {
		_, retErr := parseConfigFile(table.config)
		if retErr == nil && table.err == nil {
			continue
		}
		if retErr == nil && table.err != nil {
			t.Errorf("No error returned, expected %v", table.err)
			continue
		}
		if retErr != nil && table.err == nil {
			t.Errorf("Error returned as %v, expected nil.", retErr)
			continue
		}
		if retErr.Error() != table.err.Error() {
			t.Errorf("Error type was incorrect, got %v, want %v.", retErr.Error(), table.err)
		}
	}

}

func TestConstructArgs(t *testing.T) {
	testTable := []struct {
		filepath  string
		delimiter string
		regexp    string
		command   string
		args      string
		err       error
		sArgs     *streamArgs
	}{
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "touch",
				args:      []string{},
			},
		},
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "foo bar baz",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "myscript.sh",
				args: []string{
					"foo",
					"bar",
					"baz",
				},
			},
		},
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "foo 'bar baz'",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "touch",
				args: []string{
					"foo",
					"'bar baz'",
				},
			},
		},
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "'bar baz' foo",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "touch",
				args: []string{
					"'bar baz'",
					"foo",
				},
			},
		},
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "'filename filename2'",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "touch",
				args: []string{
					"'filename filename2'",
				},
			},
		},
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "one with several 'filename filename2'",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "touch",
				args: []string{
					"one",
					"with",
					"several",
					"'filename filename2'",
				},
			},
		},
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "foo 'bar baz' fuzz",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "touch",
				args: []string{
					"foo",
					"'bar baz'",
					"fuzz",
				},
			},
		},
		{
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "foo 'bar baz' fuzz 'groups are hard'",
			err:       nil,
			sArgs: &streamArgs{
				filepath:  "/test",
				delimiter: " ",
				regexp:    ".*",
				command:   "touch",
				args: []string{
					"foo",
					"'bar baz'",
					"fuzz",
					"'groups are hard'",
				},
			},
		},
	}

	for _, table := range testTable {
		ret, retErr := constructArgs(table.filepath, table.delimiter, table.regexp, table.command, table.args)

		if table.sArgs != nil {
			if len(table.sArgs.args) != len(ret.args) {
				t.Errorf("Returned args were not the same as expected, got %v, want %v.", ret.args, table.sArgs.args)
			} else {
				for idx := range table.sArgs.args {
					if table.sArgs.args[idx] != ret.args[idx] {
						t.Errorf("Returned args were not the same as expected, got %v, want %v.", ret.args[idx], table.sArgs.args[idx])
					}
				}
			}
		}

		if retErr == nil && table.err != nil {
			t.Errorf("No error returned, expected %v", table.err)
			continue
		}
		if retErr != nil && table.err == nil {
			t.Errorf("Error returned as %v, expected nil.", retErr)
			continue
		}
		if table.err != nil {
			if retErr.Error() != table.err.Error() {
				t.Errorf("Error type was incorrect, got %v, want %v.", retErr.Error(), table.err)
			}
		}
	}
}

func TestGetStreams(t *testing.T) {
	testTable := []struct {
		config    string
		filepath  string
		delimiter string
		regexp    string
		command   string
		args      string
		err       error
		ret       int
	}{
		{
			config:    "",
			filepath:  "/test",
			delimiter: " ",
			regexp:    ".*",
			command:   "touch",
			args:      "",
			err:       nil,
			ret:       1,
		},
		{
			config:    "../examples/streammon.conf",
			filepath:  "a",
			delimiter: "",
			regexp:    ".*",
			command:   "touch",
			args:      "",
			err:       nil,
			ret:       3,
		},
	}

	for _, table := range testTable {
		ret, retErr := getStreams(table.config, table.filepath, table.delimiter, table.regexp, table.command, table.args)
		if retErr == nil && table.err != nil {
			t.Errorf("No error returned, expected %v", table.err)
			continue
		}
		if retErr != nil && table.err == nil {
			t.Errorf("Error returned as %v, expected nil.", retErr)
			continue
		}
		if table.err != nil {
			if retErr.Error() != table.err.Error() {
				t.Errorf("Error type was incorrect, got %v, want %v.", retErr.Error(), table.err)
			}
		}
		if len(ret) != table.ret {
			t.Errorf("Expected return length of %v, got %v", table.ret, len(ret))
		}
	}
}

func TestWatchStream(t *testing.T) {
	// End to end test of the watchStream function.
	content := []byte("GET Example Request\nPOST Example Request")
	tmpfile, err := ioutil.TempFile("../examples/", "test_file")
	if err != nil {
		t.Errorf("got error opening temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(content); err != nil {
		t.Errorf("got error writing to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Errorf("got error closing temp file: %v", err)
	}

	conf := struct {
		config    string
		filepath  string
		delimiter string
		regexp    string
		command   string
		args      string
	}{
		config:    "",
		filepath:  tmpfile.Name(),
		delimiter: " ",
		regexp:    "^POST.*",
		command:   "echo",
		args:      "'#{0}'",
	}
	streams, err := getStreams(conf.config, conf.filepath, conf.delimiter, conf.regexp, conf.command, conf.args)

	if err != nil {
		t.Errorf("got error creating stream: %v", err)
	}

	var wg sync.WaitGroup
	for _, s := range streams {
		go watchStream(s, &wg)
	}

	// force close of the stream
	go func() {
		wg.Add(1)
		<-time.After(500 * time.Millisecond)
		wg.Done()
	}()

	wg.Wait()
}

func TestUsage(t *testing.T) {
	got := usage()
	if len(got) < 50 {
		t.Errorf("expected usage didnt match output. got %v", got)
	}

}
