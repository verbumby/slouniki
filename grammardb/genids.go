package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	xmlparser "github.com/tamerh/xml-stream-parser"
)

type rowT struct {
	FK     string
	Tag    string
	Lemma  string
	Suffix string
}

const DirName = "."

func main() {
	xmls, err := os.ReadDir(DirName)
	if err != nil {
		panic(err)
	}

	out, err := os.Create("ids.txt")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	totals := map[string]int{}
	rows := make([]rowT, 0, 300000)
	for _, f := range xmls {
		if !strings.HasSuffix(f.Name(), ".xml") {
			continue
		}
		filename := DirName + "/" + f.Name()

		f, err := os.Open(filename)
		if err != nil {
			panic(fmt.Errorf("open %s: %w", filename, err))
		}
		defer f.Close()

		parser := xmlparser.NewXMLParser(bufio.NewReaderSize(f, 65_536), "Paradigm")
		for paradigmXML := range parser.Stream() {
			if paradigmXML.Err != nil {
				panic(fmt.Errorf("parse paradigm: %w", paradigmXML.Err))
			}
			for _, variantXML := range paradigmXML.Childs["Variant"] {
				tag := paradigmXML.Attrs["tag"]
				if v, ok := variantXML.Attrs["tag"]; ok {
					tag += v
				}

				if tag[0] == 'K' {
					continue
				}
				lemma := variantXML.Attrs["lemma"]
				lemma = strings.ReplaceAll(lemma, "+", "")
				fk := paradigmXML.Attrs["pdgId"] + "-" + variantXML.Attrs["id"]

				totals[lemma]++
				rows = append(rows, rowT{
					FK:    fk,
					Tag:   tag,
					Lemma: lemma,
				})
			}
		}
	}

	counts := map[string]int{}
	for i := range rows {
		lemma := rows[i].Lemma
		if totals[lemma] == 1 {
			continue
		}
		counts[lemma]++
		rows[i].Suffix = strconv.FormatInt(int64(counts[lemma]), 10)
	}

	for _, row := range rows {
		fmt.Printf("%s\t%s\t%s\t%s\n", row.FK, row.Tag, row.Lemma, row.Suffix)
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", row.FK, row.Tag, row.Lemma, row.Suffix)
	}
}
