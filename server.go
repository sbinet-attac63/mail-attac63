// Copyright 2019 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

type server struct {
	http http.Server
	ln   net.Listener
}

func newServer(addr string) *server {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	srv := &server{ln: ln}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.rootHandle)
	mux.HandleFunc("/send", srv.sendHandle)

	srv.http = http.Server{
		Addr:    ln.Addr().String(),
		Handler: mux,
	}

	return srv
}

func (srv *server) run() error {
	return srv.http.Serve(srv.ln)
}

func (srv *server) rootHandle(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, page)
}

func (srv *server) sendHandle(w http.ResponseWriter, req *http.Request) {
	var msg struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	defer req.Body.Close()

	err := json.NewDecoder(req.Body).Decode(&msg)
	if err != nil {
		log.Printf("could not parse JSON request: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("subject: %q", msg.Subject)
	log.Printf("body:\n%s\n===\n", msg.Body)
}

const page = `<html>
<head>
	<title>ATTAC-63 e-mail</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css" />
	<link rel="stylesheet" href="https://www.w3schools.com/w3css/3/w3.css">
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"></script>
	
	<style>
	input[type=submit] {
		background-color: #F44336;
		padding:5px 15px;
		border:0 none;
		cursor:pointer;
		-webkit-border-radius: 5px;
		border-radius: 5px;
	}
	.flex-container {
		display: -webkit-flex;
		display: flex;
	}
	.flex-item {
		margin: 5px;
	}
	/* Safari */
	@-webkit-keyframes spin {
		0% { -webkit-transform: rotate(0deg); }
		100% { -webkit-transform: rotate(360deg); }
	}
	
	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
	</style>
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

	function sendEmail() {
		var subject = $("#mail-subject").val();
		var body = $("textarea#mail-body").val();
		$.ajax({
			url: "/send",
			method: "POST",
			data: JSON.stringify({
				"subject": subject,
				"body": body
			}),
			processData: false,
			contentType: "application/json",
			dataType: "json",
			success: function(data, status) {
				// FIXME(sbinet): report?
			},
			error: function(e) {
				alert("could not send email:\n"+e.responseText);
			}
		});
	};
	</script>

</head>

<body>
	<div id="header">
		<h2>Mail</h2>
	</div>

	<div id="content">
		<input id="mail-subject" type="text" placeholder="Sujet" size="120"><br>
		<br>
		<div>
			<textarea id="mail-body" rows="30" cols="120"></textarea>
			<br>
			<button>Pièce jointe</button>
	   </div>
	   <input type="button" onclick="sendEmail()" value="Envoyer">
	</div>
</body>
</html>
`
