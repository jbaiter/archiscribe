package main

import (
	"archiscribe/lib"
	"archiscribe/web"
)

func main() {
	lib.InitCache()
	web.Serve()
}
