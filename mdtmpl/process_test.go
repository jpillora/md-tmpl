package mdtmpl

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
		"*Updated on <!--tmpl:echo foo -->foo\n<!--/tmpl-->*",
	},
	{
		"*Pipe test <!--tmpl: echo -n abc | tr b z--><!--/tmpl-->*",
		"*Pipe test <!--tmpl:echo -n abc | tr b z -->azc<!--/tmpl-->*",
	},
	{
		"Multi <tmpl,chomp:echo foo></tmpl> and <tmpl,chomp:echo bar></tmpl>",
		"Multi <!--tmpl,chomp:echo foo -->foo<!--/tmpl--> and <!--tmpl,chomp:echo bar -->bar<!--/tmpl-->",
	},
	{
		`<!--tmpl,chomp:echo foo \&gt; bar--><!--/tmpl-->`,
		`<!--tmpl,chomp:echo foo \> bar -->foo > bar<!--/tmpl-->`,
	},
}

func TestAll(t *testing.T) {
	for i, c := range cases {
		got := Execute([]byte(c.input))
		expected := []byte(c.expected)
		if bytes.Compare(got, expected) != 0 {
			t.Errorf("Case #%d:\n====Expected====\n%s\n====Got    ====\n%s\n", i+1, expected, got)
		}
	}
}
