package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"archiscribe/lib"
	"archiscribe/web"
)

func main() {
	var logPath = flag.String("log", "", "Set path to logging file")
	var isDebug = flag.Bool("debug", false, "Enable debug mode")
	var repoPath = flag.String("repoPath", "", "Set repository path")
	flag.Parse()
	if *repoPath == "" {
		panic("repoPath must be set!")
	}
	lib.InitCache()
	if *isDebug {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else if *logPath == "" {
		log.Logger = log.Output(os.Stdout)
	} else {
		f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		log.Logger = log.Output(f)
	}
	var port int
	if *isDebug {
		port = 8083
	} else {
		port = 8080
	}
	web.Serve(port, *repoPath)
}
