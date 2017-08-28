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
			err:  errors.New(ErrFilepath),
		},
		{
			args: &streamArgs{
				filepath: "/home",
			},
			err: errors.New(ErrRegexp),
		},
		{
			args: &streamArgs{
				filepath: "/home",
				regexp:   ".*",
			},
			err: errors.New(ErrCommand),
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
