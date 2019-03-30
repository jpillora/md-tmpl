package main

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/jpillora/opts"
)

var VERSION = "0.0.0-dev"

var config = struct {
	Preview bool `help:"Enables preview mode. Displays all commands encountered."`
	Write   bool `help:"Write file instead of printing to standard out"`
	Files   []string
}{
	Preview: false,
	Write:   false,
}

func main() {
	opts.New(&config).Name("md-tmpl").Parse()
	wg := sync.WaitGroup{}
	for _, file := range config.Files {
		wg.Add(1)
		go processFile(file, &wg)
	}
	wg.Wait()
}

func processFile(file string, wg *sync.WaitGroup) {
	defer wg.Done()
	//as a safety measure, only process .md files
	if filepath.Ext(file) != ".md" {
		fmt.Printf("%s not a .md file\n", file)
		return
	}
	//read into memory
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("failed to read file: %s\n", file)
		return
	}
	s, _ := os.Stat(file)

	//process input
	commands, b := process(b)
	//preview only!
	if config.Preview {
		b := strings.Builder{}
		fmt.Fprintf(&b, "file %s contains %d commands:\n", file, len(commands))
		for i, cmd := range commands {
			fmt.Fprintf(&b, "  #%d: %s\n", i+1, cmd)
		}
		fmt.Print(b.String())
		return
	}
	//no write! print instead
	if !config.Write {
		fmt.Printf("file %s\n======\n%s\n======\n", file, b)
		return
	}
	//write back to disk
	ioutil.WriteFile(file, b, s.Mode())
	if err != nil {
		fmt.Printf("failed to write file: %s\n", file)
		return
	}
	fmt.Printf("wrote file %s (%d bytes)\n", file, len(b))
}

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

func process(input []byte) ([]string, []byte) {
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
		//instead of running the command, write the command itself
		if config.Preview {
			commands = append(commands, cmd)
			continue
		}
		//run command!
		result := run(cmd)
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

//run script by piping into bash
func run(script string) []byte {
	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader(script)
	b := &bytes.Buffer{}
	cmd.Stdout = b
	cmd.Stderr = b
	//ignore whether it failed or not
	if err := cmd.Start(); err != nil {
		log.Printf("failed to exec '%s': %s", script, err)
	}
	cmd.Wait()
	//use any output
	return bytes.TrimSuffix(b.Bytes(), exitMsg)
}

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
