package main

import (
	"fmt"
	"log"
	"os"
)

func saveTOfile(filename string, data map[string]datausage) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := fmt.Fprintln(f, "SrcIP Address\t\tRX\t\tTX"); err != nil {
		log.Fatal(err)
	}
	for _, v := range data {
		if _, err = fmt.Fprintf(f, "%v\t\t%v\t\t%v\n", v.ip, v.rx, v.tx); err != nil {
			f.Close()
			log.Fatal(err)
		}
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}

func saveTOdatabases(dnm string) {
	dbx := initDB(dnm)
	defer dbx.Close()
	createTable(dbx)
	storeItem(dbx, nmap)
	if *svtf != "false" {
		saveTOfile(*svtf, readItem(dbx))
	}
	nmap = make(map[string]datausage)
}
