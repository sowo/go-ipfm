package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

type entry struct /* sorting datausage */ {
	key   string
	value datausage
}
type byPriority []entry

var lastFlush time.Time

func init() {
	lastFlush = time.Now()
}

func (d byPriority) Len() int { return len(d) }
func (d byPriority) Less(i, j int) bool {
	if d[i].value.tstamp == d[j].value.tstamp {
		// sort by rx or tx ?
		if *srtx == "RX" {
			return d[i].value.rx < d[j].value.rx
		}
		return d[i].value.tx < d[j].value.tx
	}
	return d[i].value.tstamp < d[j].value.tstamp
}
func (d byPriority) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func saveTOdatabases(dnm string) /*save captured data to file and database */ {
	dbx := initDB(dnm) // connect to database
	defer dbx.Close()
	createTable(dbx)     // create tb if not exist
	storeItem(dbx, nmap) // save to db
	if *svtf != "false" {
		if lastFlush.Format(captureMask) != time.Now().Format(captureMask) {
			saveTOfile(*dir+"/"+*svtf, readItem(dbx, lastFlush.Format(captureMask)), lastFlush.Format(captureMask)) // save to file or not ?
		}
		saveTOfile(*dir+"/"+*svtf, readItem(dbx, time.Now().Format(captureMask)), time.Now().Format(captureMask)) // save to file or not ?
	}
	lastFlush = time.Now()
	nmap = make(map[string]datausage) // flushing ...

}

func saveTOfile(filename string, data map[string]datausage, date string) {
	f, err := os.OpenFile(filename+"."+date, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	var maxLenrx int
	slice := make(byPriority, 0, len(data))
	for key, value := range data {
		slice = append(slice, entry{key, value})
		if len(strconv.FormatUint(uint64(value.rx), 10)) > maxLenrx {
			maxLenrx = len(strconv.FormatUint(uint64(value.rx), 10))
		}
	}
	// printing header to file
	if _, err := fmt.Fprintf(f, "Go-IPFM 0.34 - Capturing Data On %v - Time: %v\n%-28v %-*v %-*v\n", *infc, time.Now().Format("2006-01-02 15:04:05"), "Source-IPV4-Add", maxLenrx+12, "RX("+*trm+")", maxLenrx+12, "TX("+*trm+")"); err != nil {
		log.Fatal(err)
	}
	// sorting data usage
	if *srfm == "descending" {
		sort.Sort(sort.Reverse(slice))
	} else {
		sort.Sort(slice)
	}

	for _, v := range slice { /*save sorted data to file ...*/
		if v.value.tstamp != date {
			continue
		}
		if _, err = fmt.Fprintf(f, "%-28v %-*.3f %-*.3f\n", v.value.ip, maxLenrx+12, float64(v.value.rx)/float64(cHz), maxLenrx+12, float64(v.value.tx)/float64(cHz)); err != nil {
			log.Fatal(err)
		}
		// fmt.Println(cHz)
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}
