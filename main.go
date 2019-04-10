package main

import (
	"sync"

	"github.com/jpillora/md-tmpl/modtmpl"
	"github.com/jpillora/opts"
)

var VERSION = "0.0.0-dev"

var root = struct {
	Files []string
}{}

func main() {
	proc := modtmpl.NewProcessor()
	opts.New(&root).Name("md-tmpl").EmbedStruct(proc).Parse()
	wg := sync.WaitGroup{}
	for _, file := range root.Files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			proc.ProcessFile(".", f)
		}(file)
	}
	wg.Wait()
}
