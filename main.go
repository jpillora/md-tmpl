package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jpillora/md-tmpl/mdtmpl"
	"github.com/jpillora/opts"
)

var VERSION = "0.0.0-dev"

var config = struct {
	Preview    bool     `help:"Enables preview mode. Displays all commands encountered."`
	Write      bool     `help:"Write file instead of printing to standard out"`
	WorkingDir string   `opts:"short=d,help=Specify the working directory for all commands (defaults to each file's dirname)"`
	Files      []string `opts:"mode=arg,min=1"`
}{
	Preview: false,
	Write:   false,
}

func main() {
	//cli
	opts.New(&config).
		Name("md-tmpl").
		Summary("Markdown template will look for 'tmpl' HTML comments in the " +
			"give files. Templates must be in the format:\n" +
			"    <!--tmpl: my-command --><!--/tmpl-->\n\n" +
			"In this case, 'my-command' would be executed via bash and the output " +
			"would be inserted in between the start and end 'tmpl' tags.").
		Repo("github.com/jpillora/md-tmpl").
		Version(VERSION).
		Parse()
	//program
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
	//preview only!
	if config.Preview {
		commands := mdtmpl.Commands(b)
		s := ""
		if len(commands) != 1 {
			s = "s"
		}
		out := fmt.Sprintf("file %s contains %d command%s:\n", file, len(commands), s)
		for i, cmd := range commands {
			out += fmt.Sprintf("  #%d: %s\n", i+1, cmd)
		}
		fmt.Print(out)
		return
	}
	//pick a working dir
	wd := config.WorkingDir
	if wd == "" {
		wd = filepath.Dir(file)
	}
	//map input to output
	b = mdtmpl.ExecuteIn(b, wd)
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
