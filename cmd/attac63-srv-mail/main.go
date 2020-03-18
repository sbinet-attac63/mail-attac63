// Copyright 2019 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"strings"

	web "github.com/toqueteos/webbrowser"
)

var (
	doDebug = flag.Bool("dbg", true, "enable debug mode")

	dbgAddrs []string
)

func main() {
	log.SetPrefix("mail-attac63: ")
	log.SetFlags(0)

	var (
		addrFlag = flag.String("addr", ":0", "[host]:port to serve")
		webFlag  = flag.Bool("web", false, "run web browser")
	)

	flag.Parse()

	srv := newServer(*addrFlag)
	log.Printf("serving mail-attac63 at %q...", srv.http.Addr)

	if *webFlag {
		go func() {
			url := srv.http.Addr
			if strings.HasPrefix(url, ":") {
				url = "localhost" + url
			}
			if !strings.HasPrefix(url, "http") {
				url = "http://" + url
			}
			log.Printf("launching web-browser to %q...", url)
			err := web.Open(url)
			if err != nil {
				log.Fatalf("could not launch web browser: %+v", err)
			}
		}()
	}

	err := srv.run()
	if err != nil {
		log.Fatalf("could not run http server: %+v", err)
	}
}
