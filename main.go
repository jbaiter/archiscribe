package main

import (
	"flag"

	"archiscribe/lib"
	"archiscribe/web"
)

func main() {
	var isDebug = flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	lib.InitCache()
	if *isDebug {
		web.Serve(8083)
	} else {
		web.Serve(8080)
	}
}
