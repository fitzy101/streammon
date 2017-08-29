package stream

import (
	"strings"
	"testing"
)

func TestPrepArgs(t *testing.T) {
	r, _ := setupRegexp("DHCPDISCOVER")
	args1 := []string{
		"ip:#{3}",
		"mac:#{5}",
		"dev:#{7}",
	}
	//args2 := []string{
	//	"ip:#{3}mac:#{5}",
	//}
	args3 := []string{
		"#{0}",
	}

	testTable := []struct {
		s    *Stream
		line string
		exp  []string
	}{
		{
			s: &Stream{
				Regexp: r,
				args:   args1,
				delim:  " ",
				fields: parseFields(args1),
			},
			line: `DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1`,
			exp: []string{
				"ip:192.168.127.3",
				"mac:61:7c:db:fb:45:5e",
				"dev:br1",
			},
		},
		{
			s: &Stream{
				Regexp: r,
				args:   args3,
				delim:  " ",
				fields: parseFields(args3),
			},
			line: `DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1`,
			exp:  []string{"DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1"},
		},
	}

	for _, test := range testTable {
		resp := prepArgs(test.line, test.s)
		if len(resp) != len(test.exp) {
			t.Errorf("response was different length, expected %v, got %v", test.exp, resp)
		}
		for idx, _ := range resp {
			if strings.Compare(resp[idx], test.exp[idx]) != 0 {
				t.Errorf("response strings were different, expected %v, got %v", test.exp[idx], resp[idx])
			}
		}
	}
}
