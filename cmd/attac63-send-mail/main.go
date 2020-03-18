// Copyright 2020 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/sbinet-attac63/mail-attac63/auth"
	gomail "gopkg.in/gomail.v2"
)

func main() {
	log.SetPrefix("mail: ")
	log.SetFlags(0)

	var (
		md  = flag.String("md", "input.md", "path to email content")
		lst = flag.String("db", "list.csv", "path to email addresses")
		dbg = flag.Bool("dbg", true, "enable debug mode")
	)

	flag.Parse()

	raw, err := ioutil.ReadFile(*md)
	if err != nil {
		log.Fatalf("could not open email content: %+v", err)
	}

	title := ""

	renderHookDropTitle := func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		switch node := node.(type) {
		case *ast.Document:
			if entering {
				title = string(node.Children[0].(*ast.Heading).Children[0].(*ast.Text).Literal)
				ast.RemoveFromTree(node.Children[0])
			}
		}
		return ast.GoToNext, false
	}

	opts := html.RendererOptions{
		Flags:          html.CommonFlags,
		RenderNodeHook: renderHookDropTitle,
	}
	renderer := html.NewRenderer(opts)

	html := markdown.ToHTML(raw, nil, renderer)

	log.Printf("=== %s ===\n%s\n===\n", title, html)

	if !*dbg {
		msg := newEmail(title, html, *lst, flag.Args())
		err = msg.send()
		if err != nil {
			log.Fatalf("could not send emails: %+v", err)
		}
	}
}

type Email struct {
	subject string
	content string

	rcpts   []string
	attachs []string
}

func newEmail(subject string, content []byte, emails string, attachs []string) Email {
	msg := Email{
		subject: subject,
		content: string(content),
		rcpts:   loadAddresses(emails),
		attachs: attachs,
	}

	return msg
}

func (msg Email) send() error {
	var (
		err error
		ok  = true
	)

	for i, rcpt := range msg.rcpts {
		log.Printf("sending mail %d/%d (%q)...", i+1, len(msg.rcpts), rcpt)
		m := gomail.NewMessage()
		m.SetHeader("From", auth.Usr)
		m.SetHeader("Bcc", rcpt)
		m.SetHeader("Subject", msg.subject)
		m.SetBody("text/html", msg.content)
		for _, file := range msg.attachs {
			m.Attach(file)
		}

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

func loadAddresses(fname string) []string {
	var addrs []string

	f, err := os.Open(fname)
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
	for _, rec := range recs {
		addrs = append(addrs, rec[2])
	}
	return addrs
}
