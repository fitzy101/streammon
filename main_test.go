package main

import (
	"errors"
	"testing"
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
					"args":"publish key 'other value''"
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
