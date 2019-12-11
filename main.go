// Copyright 2019 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	web "github.com/toqueteos/webbrowser"
)

func main() {
	log.SetPrefix("mail-attac63: ")
	log.SetFlags(0)

	var (
		addrFlag = flag.String("addr", ":8080", "[host]:port to serve")
		webFlag  = flag.Bool("web", false, "run web browser")
	)

	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandle)

	log.Printf("serving mail-attac63 at %q...", *addrFlag)

	if *webFlag {
		go func() {
			url := *addrFlag
			if strings.HasPrefix(url, ":") {
				url = "localhost" + url
			}
			if !strings.HasPrefix(url, "http") {
				url = "http://" + url
			}
			log.Printf("launching web-browser to %q...", url)
			time.Sleep(2 * time.Second)
			err := web.Open(url)
			if err != nil {
				log.Fatalf("could not launch web browser: %+v", err)
			}
		}()
	}

	err := http.ListenAndServe(*addrFlag, mux)
	if err != nil {
		log.Fatalf("could not run http server: %+v", err)
	}
}

func rootHandle(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, page)
}

const page = `
<html>
	<head>
		<title>ATTAC-63 e-mail</title>
		<script type="text/javascript">
		var sock = null;

		function update() {
		};

		window.onload = function() {
			sock = new WebSocket("ws://"+location.host+"/data");

			sock.onmessage = function(event) {
				var data = JSON.parse(event.data);
				//console.log("data: "+JSON.stringify(data));
				update();
			};
		};

		</script>

		<style>
		</style>
	</head>

	<body>
		<div id="header">
			<h2>Mail</h2>
		</div>

		<div id="content">
			<div id="mail-subject" class="mail-subject"></div>
			<br>
			<div id="mail-body" class="mail-body"></div>
		</div>
	</body>
</html>
`
