// Copyright 2020 The mail-attac63 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sbinet-attac63/mail-attac63/internal/xls"
)

func main() {
	oname := flag.String("o", "testdata/liste.csv", "name of output CSV file")

	flag.Parse()

	fname := flag.Arg(0)
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("could not open %q: %+v", fname, err)
	}
	defer f.Close()

	out, err := os.Create(*oname)
	if err != nil {
		log.Fatalf("could not create %q: %+v", *oname, err)
	}
	defer out.Close()

	csv := csv.NewWriter(out)
	defer csv.Flush()

	wb, err := xls.OpenReader(f, "utf-8")
	if err != nil {
		log.Fatalf("could not open workbook: %+v", err)
	}

	log.Printf("numsheets: %d", wb.NumSheets())

	p := wb.GetSheet(0)
	if p == nil {
		log.Fatalf("could not open sheet #0")
	}

	log.Printf("sheet: %q, rows=%d", p.Name, p.MaxRow)
	for i := 0; i <= int(p.MaxRow); i++ {
		row := p.Row(i)
		if row == nil {
			log.Printf("could not get row #%d", i)
			continue
		}
		if row.Col(row.FirstCol()) == "" {
			continue
		}

		var (
			id      = strings.TrimSpace(row.Col(0))
			name    = strings.TrimSpace(row.Col(1))
			surname = strings.TrimSpace(row.Col(2))
			email   = strings.TrimSpace(row.Col(8))
		)
		if _, ok := strconv.Atoi(id); ok != nil {
			continue
		}

		if grps := reMail.FindStringSubmatch(email); grps != nil {
			email = grps[1]
		}
		email = strings.Replace(email, ",", ".", -1)

		if email == "" {
			continue
		}

		err = csv.Write([]string{surname, name, email})
		if err != nil {
			log.Fatalf("could not convert row %d to CSV: %+v", i, err)
		}
	}

	err = csv.Error()
	if err != nil {
		log.Fatalf("could not process CSV: %+v", err)
	}

	csv.Flush()

	err = out.Close()
	if err != nil {
		log.Fatalf("could not close output CSV file: %+v", err)
	}
}

var reMail = regexp.MustCompile(".*?[(]mailto:(?P<email>.*?)[)]")
