// Copyright 2019 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	gomail "gopkg.in/gomail.v2"
)

var (
	attacUsr = ""
	attacPwd = ""
	attacSrv = ""
	attacLst = []string{"sbinet@pm.me"}
)

type Mail struct {
	subject string
	body    string
	files   []string

	dir string // tmp dir holding attachments
}

func newMail() (*Mail, error) {
	tmp, err := ioutil.TempDir("", "mail-attac63-dir-")
	if err != nil {
		return nil, fmt.Errorf("could not create tmpdir: %w", err)
	}

	return &Mail{dir: tmp}, nil
}

func (m *Mail) Close() error {
	return os.RemoveAll(m.dir)
}

func (m *Mail) attach(name string, f io.Reader) error {
	fname := fmt.Sprintf("file-%03d", len(m.files))
	o, err := os.Create(filepath.Join(m.dir, fname))
	if err != nil {
		return fmt.Errorf("could not add attachment %q: %w", name, err)
	}
	defer o.Close()

	_, err = io.Copy(o, f)
	if err != nil {
		return fmt.Errorf("could not copy attachment %q: %w", name, err)
	}

	err = o.Close()
	if err != nil {
		return fmt.Errorf("could not close attachment %q: %w", name, err)
	}

	m.files = append(m.files, name)
	return nil
}

func (srv *server) send() error {
	defer func() {
		err := srv.mail.Close()
		if err != nil {
			log.Printf("could not cleanup e-mail: %+v", err)
		}
		srv.mail = nil
	}()

	for rcpts := range srv.chunks(20, srv.rcpts) {
		log.Printf("sending to %q", rcpts)
		mail := gomail.NewMessage()
		mail.SetHeader("From", attacUsr)
		mail.SetHeader("Subject", srv.mail.subject)
		mail.SetBody("text/html", srv.mail.body)
		for i, name := range srv.mail.files {
			mail.Attach(
				filepath.Join(srv.mail.dir, fmt.Sprintf("file-%03d", i)),
				gomail.Rename(name),
			)
		}

		mail.SetHeader("Bcc", rcpts...)

		err := srv.dial.DialAndSend(mail)
		if err != nil {
			return fmt.Errorf("could not send mail to %q: %w", rcpts, err)
		}
	}
	return nil
}

func (srv *server) chunks(sz int, vs []string) chan []string {
	ch := make(chan []string)
	go func() {
		defer close(ch)
		if len(vs) <= sz {
			ch <- vs
			return
		}

		for i := 0; i < len(vs); i += sz {
			beg := i
			end := i + sz
			if end > len(vs) {
				end = len(vs)
			}
			ch <- vs[beg:end]
		}
		return
	}()
	return ch
}
