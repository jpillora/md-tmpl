package mdtmpl

import (
	"bytes"
	"html"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

var start = regexp.MustCompile(
	//open
	`<(!--)?tmpl` +
		//optional flags
		`(` +
		`(,\w+(=\w+)?)*` +
		`)` +
		//command
		`:([^>]+?)` +
		//close
		`(--)?>`)
var end = regexp.MustCompile(
	//open and close
	`<(!--)?/tmpl(--)?>`)

//Commands returns the command strings, and
//does NOT execute them.
func Commands(md []byte) []string {
	cmds, _ := process(md, "", true)
	return cmds
}

//Execute the all commands found in the given file
//and store the result in between the comment tags:
//    <!--tmpl: my-command -->
//    some output of my-command goes here
//    <!--/tmpl-->
//It does not return an error. Both successful and
//failing commands will return their output.
func Execute(md []byte) []byte {
	_, out := process(md, "", false)
	return out
}

//ExecuteIn performs an Execute in the specified
//working directory.
func ExecuteIn(md []byte, workingDir string) []byte {
	_, out := process(md, workingDir, false)
	return out
}

func process(input []byte, workingDir string, commandsOnly bool) ([]string, []byte) {
	commands := []string{}
	output := bytes.Buffer{}
	curr := input
	for {
		//match next template open
		m := start.FindSubmatchIndex(curr)
		if m == nil {
			output.Write(curr)
			break
		}
		strs := getStrings(curr, m)
		//match result contains pairs
		//[all 0 1] ... [options 4 5] [last op 6 7] [cmd 10 11] ...
		pre := curr[:m[0]]
		cmd := strs[5]
		//check opts
		code := ""
		chomp := false
		if flags := strs[2]; flags != "" {
			//trim comma prefix
			flags = flags[1:]
			//loop each
			for _, o := range strings.Split(flags, ",") {
				kv := strings.Split(o, "=")
				k := kv[0]
				v := ""
				if len(kv) == 2 {
					v = kv[1]
				}
				switch k {
				case "code":
					code = v
					if code == "" {
						code = "plain"
					}
					chomp = true //code forces newlines, so just make sure there's only one
				case "chomp":
					chomp = true
				default:
					log.Printf("[mdtmpl] unknown option '%s'", k)
				}
			}
		}
		//keep offset of end of openning tag
		o := m[1]
		//match next template close, from offset
		m = end.FindSubmatchIndex(curr[o:])
		if m == nil {
			output.Write(curr)
			break
		}
		//set input to everything after end tag
		curr = curr[o+m[1]:]
		//html entity decode command
		cmd = strings.TrimSpace(html.UnescapeString(string(cmd)))
		//collect commands
		commands = append(commands, cmd)
		//dont execute commands
		if commandsOnly {
			continue
		}
		//run command!
		c := exec.Command("bash")
		if workingDir != "" {
			c.Dir = workingDir
		}
		c.Stdin = strings.NewReader(cmd)
		out, err := c.CombinedOutput()
		//ignore whether it failed or not
		if err != nil {
			log.Printf("failed to exec '%s': %s", cmd, err)
		}
		result := bytes.TrimSuffix(out, exitMsg)
		//trim *last* newline
		if chomp && bytes.HasSuffix(result, []byte("\n")) {
			result = result[:len(result)-1] //newline is 13 => 1 byte
		}
		//wrap in code block
		if code != "" {
			result = append([]byte("\n``` "+code+" \n"), result...)
			result = append(result, []byte("\n```\n")...)
		}
		//replace last result with new result
		output.Write(pre)
		output.WriteString("<!--tmpl")
		if chomp {
			output.WriteString(",chomp")
		}
		if code != "" {
			output.WriteString(",code=")
			output.WriteString(code)
		}
		output.WriteRune(':')
		output.WriteString(cmd)
		output.WriteString(" -->")
		output.Write(result)
		output.WriteString("<!--/tmpl-->")
	}
	return commands, output.Bytes()
}

var exitMsg = []byte("exit status 1\n")

func getStrings(b []byte, indexPairs []int) []string {
	num := len(indexPairs) / 2
	strs := make([]string, num)
	for p := 0; p < num; p++ {
		pi := p * 2
		pj := pi + 1
		start := indexPairs[pi]
		end := indexPairs[pj]
		if start == -1 {
			continue
		}
		str := string(b[start:end])
		strs[p] = str
	}
	return strs
}
