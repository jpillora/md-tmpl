package main

import (
	"bytes"
	"testing"
)

var cases = []struct {
	input, expected string
}{
	{
		"abc",
		"abc",
	},
	{
		"*Updated on <tmpl:echo foo></tmpl>*",
		"*Updated on <tmpl:echo foo>foo\n</tmpl>*",
	},
	{
		"*Pipe test <tmpl: echo -n abc | tr b z></tmpl>*",
		"*Pipe test <tmpl: echo -n abc | tr b z>azc</tmpl>*",
	},
	{
		"Multi <tmpl,chomp:echo foo></tmpl> and <tmpl,chomp:echo bar></tmpl>",
		"Multi <tmpl,chomp:echo foo>foo</tmpl> and <tmpl,chomp:echo bar>bar</tmpl>",
	},
	{
		"<tmpl,chomp:echo foo &gt; bar></tmpl>",
		"<tmpl,chomp:echo foo &gt; bar>foo > bar</tmpl>",
	},
}

func TestAll(t *testing.T) {
	for i, c := range cases {
		out := process([]byte(c.input))
		expected := []byte(c.expected)
		if bytes.Compare(out, expected) != 0 {
			t.Errorf("Case #%d: Expected and got\n====\n%s\n====\n%s\n", i+1, expected, out)
		}
	}
}
