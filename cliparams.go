package main

import (
	"flag"
)

var (
	listen = ":8080"
)

func parseFlags() {
	flag.StringVar(&listen, "listen", listen, "Address to listen as HTTP server")
	flag.Parse()
}
