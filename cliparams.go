package main

import (
	"flag"
	"strings"
)

var (
	paths  = arrayFlags{}
	listen = ":8080"
)

func parseFlags() {
	flag.Var(&paths, "path", "URL Path where the Webhook requests will come in (can be repeated multiple times)")
	flag.StringVar(&listen, "listen", listen, "Address to listen as HTTP server")
	flag.Parse()
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "arrayFlags"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}
