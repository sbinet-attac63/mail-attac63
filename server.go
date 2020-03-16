// Copyright 2019 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

type server struct {
	http  http.Server
	ln    net.Listener
	mail  *Mail
	dial  *gomail.Dialer
	rcpts []string
	vers  string
}

func newServer(addr string) *server {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	mail, err := newMail()
	if err != nil {
		panic(err)
	}

	srv := &server{
		ln:    ln,
		rcpts: attacLst,
		mail:  mail,
		dial:  gomail.NewDialer(attacSrv, 465, attacUsr, attacPwd),
		vers:  Version,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.rootHandle)
	mux.HandleFunc("/send", srv.sendHandle)
	mux.HandleFunc("/attach", srv.attachHandle)

	srv.http = http.Server{
		Addr:    ln.Addr().String(),
		Handler: mux,
	}
	srv.mail = mail

	return srv
}

func (srv *server) run() error {
	return srv.http.Serve(srv.ln)
}

var (
	tmpl = template.Must(template.New("main").Parse(page))
)

func (srv *server) rootHandle(w http.ResponseWriter, req *http.Request) {
	err := tmpl.Execute(w, struct {
		Version string
	}{Version})
	if err != nil {
		log.Printf("could not execute main template: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	if srv.mail == nil {
		srv.mail, err = newMail()
		if err != nil {
			log.Printf("could not create mailer: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	srv.mail.subject = msg.Subject
	srv.mail.body = strings.ReplaceAll(
		template.HTMLEscapeString(msg.Body),
		"\n",
		"<br>\n",
	)

	err = srv.send()
	if err != nil {
		log.Printf("could not send email: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"ok"})
	if err != nil {
		log.Printf("could not send ack: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("email sent")
}

func (srv *server) attachHandle(w http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(50 * 1024 * 1024)
	if err != nil {
		log.Printf("could not parse multipart form: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f, hdr, err := req.FormFile("filedata")
	if err != nil {
		log.Printf("error form-file: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if srv.mail == nil {
		srv.mail, err = newMail()
		if err != nil {
			log.Printf("could not create mailer: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = srv.mail.attach(hdr.Filename, f)
	if err != nil {
		log.Printf("could not create attach-file %q: %+v", hdr.Filename, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

const page = `<html>
<head>
	<title>ATTAC-63 e-mail</title>

	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css" />
	<link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"></script>

	<style>
	input[type=file] {
		display: none;
	}
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
	.app-file-upload {
		color: white;
		background-color: #0091EA;
		padding:5px 15px;
		border:0 none;
		cursor:pointer;
		-webkit-border-radius: 5px;
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
	"use strict"
	var sock = null;

	function update() {
	};

	window.onload = function() {
	//	sock = new WebSocket("ws://"+location.host+"/data");

	//	sock.onmessage = function(event) {
	//		var data = JSON.parse(event.data);
	//		//console.log("data: "+JSON.stringify(data));
	//		update();
	//	};
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
				document.getElementById("mail-ok").style.display="block";
			},
			error: function(e) {
				alert("could not send email:\n"+e.responseText);
			}
		});
	};

	function attach() {
		$("#mail-attach").click();
	}

	$(function () {
		document.getElementById("mail-attach").onchange = function() {
			console.log("attached");
			var data = new FormData($("#mail-upload-form")[0]);
			var fname = data.get("filedata").name;
			console.log("file: ["+fname+"]");
			$.ajax({
				url: "/attach",
				method: "POST",
				data: data,
				processData: false,
				contentType: false,
				success: function(data, status) {
					$("#mail-attach-list").append(
					'<li class="w3-display-container">'+fname+'<span onclick="this.parentElement.style.display=\'none\'" class="w3-button w3-transparent w3-display-right">&times;</span></li>'
					);
				},
				error: function(er){
					alert("upload failed: "+er);
				}
			});
		}
	});

	</script>

</head>

<body>

<!-- Sidebar -->
<div id="app-sidebar" class="w3-sidebar w3-bar-block w3-card-4 w3-light-grey" style="width:25%">
	<div class="w3-bar-item w3-card-2 w3-black">
		<h2>ATTAC63</h2>
	</div>
	<div class="w3-bar-item">
		<br>
		<pre>Version: {{.Version}}</pre>
		<br>
	</div>
</div>

<!-- Page Content -->
<div style="margin-left:25%; height:100%" class="w3-container" id="app-container">
	<br>
	<div class="w3-container w3-content w3-card w3-border w3-border-grey w3-border-round" style="width:100%" id="app-display">
		<div class="w3-container w3-content w3-indigo">
			<h2>E-mail</h2>
		</div>
		<div>
			<br>
			<input class="w3-input w3-border" id="mail-subject" type="text" placeholder="Sujet">
			<br>
			<br>
			
			<textarea class="w3-input w3-border" id="mail-body" rows="15" cols="110"></textarea>

			<ul id="mail-attach-list" class="w3-ul w3-card-4">
			</ul>
			
			<button class="w3-button w3-indigo w3-xlarge w3-padding-large" onclick="attach()"><i class="fa fa-paperclip"></i></button>
			<button class="w3-button w3-red w3-right w3-xlarge w3-padding-large" onclick="sendEmail()"><i class="fa fa-paper-plane"></i></button>
			<br>
			<br>
			<form id="mail-upload-form" enctype="multipart/form-data" action="/attach" method="post">
			<input type="file" id="mail-attach" name="filedata"/>
			</form>

		</div>
	</div>

  <div id="mail-ok" class="w3-modal">
    <div class="w3-modal-content">
      <div class="w3-container w3-green w3-center w3-justify">
        <span onclick="document.getElementById('mail-ok').style.display='none'" class="w3-button w3-display-topright">&times;</span>
        <p>Success!</p>
      </div>
    </div>
  </div>

</div>

</body>
</html>
`
