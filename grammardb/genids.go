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
	Suffix *string
}

const DirName = "."

func load23ids() map[string]rowT {
	ids23f, err := os.Open("./ids-2025.txt")
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(ids23f)

	ans := map[string]rowT{}
	for s.Scan() {
		ps := strings.Split(s.Text(), "\t")
		id, tag, lemma, suffix := ps[0], ps[1], ps[2], ps[3]
		ans[id] = rowT{FK: id, Tag: tag, Lemma: lemma, Suffix: &suffix}
	}

	return ans
}

func main() {
	ids23 := load23ids()

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
				row := rowT{
					FK:    fk,
					Tag:   tag,
					Lemma: lemma,
				}

				if row23, ok := ids23[row.FK]; ok {
					if row.Tag != row23.Tag || row.Lemma != row23.Lemma {
						// fmt.Printf("tag or lemma do not match: %v vs %v\n", row, row23)
					} else {
						row.Suffix = row23.Suffix
					}
				}
				rows = append(rows, row)
			}
		}
	}

	suffixes := map[string]bool{}
	for i := range rows {
		if rows[i].Suffix == nil {
			continue
		}
		suff := *(rows[i].Suffix)
		if suff == "" {
			suff = "1"
		}
		suffixes[rows[i].Lemma+"-"+suff] = true
	}

	for i := range rows {
		n := 1
		for rows[i].Suffix == nil {
			ns := strconv.FormatInt(int64(n), 10)
			id := rows[i].Lemma + "-" + ns
			if !suffixes[id] {
				suffixes[id] = true
				if ns == "1" {
					ns = ""
				}
				rows[i].Suffix = &ns
			}
			n++
		}
	}

	for _, row := range rows {
		fmt.Printf("%s\t%s\t%s\t%s\n", row.FK, row.Tag, row.Lemma, *row.Suffix)
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", row.FK, row.Tag, row.Lemma, *row.Suffix)
	}
}
