package main

import (
	"flag"

	"archiscribe/lib"
	"archiscribe/web"
)

func main() {
	var isDebug = flag.Bool("debug", false, "Enable debug mode")
	var repoPath = flag.String("repoPath", "", "Set repository path")
	flag.Parse()
	if *repoPath == "" {
		panic("repoPath must be set!")
	}
	lib.InitCache()
	if *isDebug {
		web.Serve(8083, *repoPath)
	} else {
		web.Serve(8080, *repoPath)
	}
}
