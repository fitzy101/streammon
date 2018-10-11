package stream

import (
	"errors"
	"strings"
	"testing"
)

func TestPrepArgs(t *testing.T) {
	r, _ := setupRegexp("DHCPDISCOVER")

	testTable := []struct {
		s    *Stream
		line string
		exp  []string
	}{
		{
			s: &Stream{
				Regexp: r,
				args:   []string{},
				delim:  " ",
				fields: parseFields([]string{}),
			},
			line: `DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1`,
			exp:  []string{},
		},
		{
			s: &Stream{
				Regexp: r,
				args: []string{
					"ip:#{3}",
					"mac:#{5}",
					"dev:#{7}",
				},
				delim: " ",
				fields: parseFields(
					[]string{
						"ip:#{3}",
						"mac:#{5}",
						"dev:#{7}",
					},
				),
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
				args: []string{
					"ip:#{3},mac:#{5}",
				},
				delim: " ",
				fields: parseFields(
					[]string{
						"ip:#{3},mac:#{5}",
					},
				),
			},
			line: `DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1`,
			exp: []string{
				"ip:192.168.127.3,mac:61:7c:db:fb:45:5e",
			},
		},
		{
			s: &Stream{
				Regexp: r,
				args: []string{
					"#{0}",
				},
				delim: " ",
				fields: parseFields(
					[]string{
						"#{0}",
					},
				),
			},
			line: `DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1`,
			exp:  []string{"DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1"},
		},
		{
			s: &Stream{
				Regexp: r,
				args: []string{
					"ip:#{3},mac:#{5},dev:#{7}",
				},
				delim: " ",
				fields: parseFields(
					[]string{
						"ip:#{3},mac:#{5},dev:#{7}",
					},
				),
			},
			line: `DHCPREQUEST for 192.168.127.3 from 61:7c:db:fb:45:5e via br1`,
			exp: []string{
				"ip:192.168.127.3,mac:61:7c:db:fb:45:5e,dev:br1",
			},
		},
	}

	for _, test := range testTable {
		resp := prepArgs(test.line, test.s)
		if len(resp) != len(test.exp) {
			t.Errorf("response was different length, expected %v, got %v", test.exp, resp)
		}
		for idx := range resp {
			if strings.Compare(resp[idx], test.exp[idx]) != 0 {
				t.Errorf("response strings were different, expected %v, got %v", test.exp[idx], resp[idx])
			}
		}
	}
}

func TestParseFields(t *testing.T) {
	testTable := []struct {
		args []string
		exp  []int
	}{
		{
			args: []string{
				"ip:#{3}",
				"mac:#{5}",
				"dev:#{7}",
			},
			exp: []int{3, 5, 7},
		},
		{
			args: []string{
				"ip:#{3},mac:#{5}",
			},
			exp: []int{3, 5},
		},
		{
			args: []string{
				"#{0}",
			},
			exp: []int{0},
		},
	}

	for _, test := range testTable {
		resp := parseFields(test.args)
		if len(resp) != len(test.exp) {
			t.Errorf("response was different length, expected %v, got %v", test.exp, resp)
		}
		for idx := range resp {
			if resp[idx] != test.exp[idx] {
				t.Errorf("response ints were different, expected %v, got %v", test.exp[idx], resp[idx])
			}
		}
	}

}

func TestInsertField(t *testing.T) {
	testTable := []struct {
		initial string
		replace string
		field   int
		exp     string
	}{
		{
			initial: "ip-address:#{1}",
			replace: "10.121.0.1",
			field:   1,
			exp:     "ip-address:10.121.0.1",
		},
		{
			initial: "ip-address:#{2}",
			replace: "10.121.0.1",
			field:   1,
			exp:     "ip-address:#{2}",
		},
		{
			initial: "ip-address",
			replace: "10.121.0.1",
			field:   1,
			exp:     "ip-address",
		},
	}

	for _, test := range testTable {
		resp := insertField(test.initial, test.replace, test.field)
		if strings.Compare(resp, test.exp) != 0 {

			t.Errorf("response string was incorrect, expected %v, got %v", test.exp, resp)
		}
	}
}

func TestSetupRegexp(t *testing.T) {
	testTable := []struct {
		pattern string
		exp     error
	}{
		{
			pattern: `(`,
			exp:     errors.New("error parsing regexp: missing closing ): `(`"),
		},
		{
			pattern: `.*`,
		},
		{
			pattern: `abc`,
		},
		{
			pattern: `[0-9a-zA-Z]`,
		},
	}

	for _, test := range testTable {
		_, resp := setupRegexp(test.pattern)
		if test.exp != nil && resp == nil {
			t.Errorf("errors not returned, expected %v, got nil", test.exp.Error())
		}
		if test.exp == nil && resp != nil {
			t.Errorf("error returned, expected nil, got %v", resp.Error())
		}
		if test.exp != nil && resp != nil {
			if strings.Compare(test.exp.Error(), resp.Error()) != 0 {
				t.Errorf("unexpected error returned, expected %v, got %v", test.exp.Error(), resp.Error())
			}
		}
	}
}
