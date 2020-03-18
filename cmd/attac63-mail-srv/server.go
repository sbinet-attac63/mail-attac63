// Copyright 2019 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/sbinet-attac63/mail-attac63/auth"
	gomail "gopkg.in/gomail.v2"
)

type server struct {
	http http.Server
	ln   net.Listener
	dbg  bool
}

func newServer(addr string) *server {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	srv := &server{ln: ln, dbg: *doDebug}

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

	err = srv.send(msg.Subject, msg.Body)
	if err != nil {
		log.Printf("could not send emails: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(nil)
	return
}

func (srv *server) send(subject, body string) error {
	var (
		rcpts = srv.loadAddresses()
		m     = gomail.NewMessage()
		err   error
		ok    = true
	)

	body = htmlify(body)

	for _, rcpt := range rcpts {
		m.SetHeader("From", auth.Usr)
		m.SetHeader("Bcc", rcpt)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", body)

		d := gomail.NewDialer(auth.Srv, 465, auth.Usr, auth.Pwd)
		e := d.DialAndSend(m)
		if e != nil {
			ok = false
			log.Printf("could not send email to %q: %+v", rcpt, e)
			if err == nil {
				err = e
			}
		}
	}

	if !ok {
		return fmt.Errorf("could not send emails: %w", err)
	}
	log.Printf("emails sent successfully")

	return nil
}

func (srv *server) loadAddresses() []string {
	var addrs []string
	if srv.dbg {
		return dbgAddrs
	}

	f, err := os.Open("./liste.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ','

	recs, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	for i, rec := range recs {
		if i == 0 {
			continue
		}
		addrs = append(addrs, rec[2])
	}
	return addrs
}

func htmlify(s string) string {
	s = strings.Replace(s, "\n", "<br>\n", -1)
	return s
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
