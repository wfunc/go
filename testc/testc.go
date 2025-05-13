package main

import (
	"github.com/wfunc/go/transport"
	"github.com/wfunc/web"
)

func main() {
	forward, _ := transport.NewTransportH("tcp://192.168.1.1:80")
	web.Shared.Handle("/", forward)
	web.ListenAndServe(":9322")
}
