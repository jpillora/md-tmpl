package modtmpl

import (
	"bytes"
	"fmt"
	"testing"
)

var cases = []struct {
	name, input, expected string
}{
	{
		"abc",
		"abc",
		"abc",
	},
	{
		"echo",
		"*Updated on <!--tmpl:echo foo--><!--/tmpl-->*",
		"*Updated on <!--tmpl:echo foo-->foo\n<!--/tmpl-->*",
	},
	{
		"echo_pipe",
		"*Pipe test <!--tmpl:echo -n abc | tr b z--><!--/tmpl-->*",
		"*Pipe test <!--tmpl:echo -n abc | tr b z-->azc<!--/tmpl-->*",
	},
	{
		"multi",
		"Multi <!--tmpl,chomp:echo foo--><!--/tmpl--> and <!--tmpl,chomp:echo bar--><!--/tmpl-->",
		"Multi <!--tmpl,chomp:echo foo-->foo<!--/tmpl--> and <!--tmpl,chomp:echo bar-->bar<!--/tmpl-->",
	},
	{
		"html_unesc",
		`<!--tmpl,chomp:echo foo \&gt; bar--><!--/tmpl-->`,
		`<!--tmpl,chomp:echo foo \&gt; bar-->foo > bar<!--/tmpl-->`,
	},
}

func TestAll(t *testing.T) {
	for i, c := range cases {
		_, out := process(false, []byte(c.input))
		expected := []byte(c.expected)
		if bytes.Compare(out, expected) != 0 {
			t.Errorf("**(%d) %s** : Expected and got\n====\n%s\n====\n%s\n", i+1, c.name, expected, out)
		}
	}
}

func TestRun(t *testing.T) {
	r := run("date")
	fmt.Print(string(r))
}
