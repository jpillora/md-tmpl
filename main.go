package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
)

var VERSION = "0.0.0-dev"
var preview = flag.Bool("preview", false, "")
var p = flag.Bool("p", false, "preview mode alias")

func main() {
	flag.Usage = func() {
		fmt.Printf(`
	Usage: md-tmpl [options] markdown-files

	Version: ` + VERSION + `

	Options:
	  --preview, -p, Enables preview mode. Displays all commands
	  encountered and performs no writes.

	Read more:
	  https://github.com/jpillora/md-tmpl

`)
	}
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		return
	}

	if *p {
		*preview = true
	}
	if *preview {
		fmt.Printf("Preview mode\n")
	}

	for _, file := range files {
		processFile(file)
	}
}

func processFile(file string) {
	fmt.Printf("%s -", file)
	//as a safety measure, only process .md files
	if filepath.Ext(file) != ".md" {
		fmt.Printf(" not a .md file\n")
		return
	}
	//read into memory
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf(" failed to read file\n")
		return
	}
	//process input
	b = process(b)

	//no writes!
	if *preview {
		return
	}

	//write back to disk
	ioutil.WriteFile(file, b, 0644)
	if err != nil {
		fmt.Printf(" failed to write file\n")
		return
	}

	fmt.Printf(" success\n")
}

//HACK: two "code"s allows it to occur at either side
//	(this will not work for any more options)
var start = regexp.MustCompile(`<tmpl(,code)?(,chomp)?(,code)?:([^>]+)>`)
var end = regexp.MustCompile(`</tmpl>`)

func process(input []byte) []byte {

	var output []byte

	for {
		//match next template open
		m := start.FindSubmatchIndex(input)
		if m == nil {
			output = append(output, input...)
			break
		}
		//match result contains pairs
		//[all 0 1] [code 2 3] [chomp 4 5] [code 6 7]  [cmd 8 9]
		pre := input[:m[1]]
		code := m[2] > 0 || m[6] > 0
		chomp := code || m[4] > 0
		cmd := input[m[8]:m[9]]

		//match next template close, from offset
		o := m[1]
		m = end.FindSubmatchIndex(input[o:])
		if m == nil {
			output = append(output, input...)
			break
		}

		end := input[o+m[0] : o+m[1]]
		//safe to trim input
		input = input[o+m[1]:]

		//display command and skip
		if *preview {
			fmt.Printf("  %s\n", cmd)
			continue
		}

		//run command!
		result := run(cmd)

		//swap in gt symbols
		result = bytes.Replace(result, []byte("&gt;"), []byte(">"), -1)

		//trim newline
		if chomp {
			result = bytes.TrimRight(result, "\n")
		}
		//wrap in code block
		if code {
			result = append([]byte("\n```\n"), result...)
			result = append(result, []byte("\n```\n")...)
		}

		//replace last result with new result
		output = append(output, pre...)
		output = append(output, result...)
		output = append(output, end...)

	}

	return output
}

//run script by piping into bash
func run(script []byte) []byte {
	cmd := exec.Command("bash")

	//get all the pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	//pipe script into shell
	stdin.Write(script)
	stdin.Close()

	//roll-your-own cmd.OutputCombined
	b := &bytes.Buffer{}
	go func() {
		io.Copy(b, stdout)
	}()
	go func() {
		io.Copy(b, stderr)
	}()

	//ignore whether it failed or not
	cmd.Run()

	return b.Bytes()
}
