package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type datausage struct {
	ip     string
	tx, rx uint
	tstamp string
}

func initDB(filepath string) *sql.DB {
	// initial database
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func createTable(db *sql.DB) {
	// create table if not exists
	sqlTable := `
	CREATE TABLE IF NOT EXISTS items(
		IP TEXT NOT NULL,
		RX TEXT,
		TX TEXT,
		TIME TEXT,
		PRIMARY KEY (IP, TIME)
	);
	`

	if _, err := db.Exec(sqlTable); err != nil {
		log.Fatal(err)
	}

}

func storeItem(db *sql.DB, items map[string]datausage) {
	// store items in database
	sqlAdditem := `
	INSERT OR REPLACE INTO items(
		IP,
		RX,
		TX,
		TIME
	) values(?, ?, ?, ?)
	`

	sqlSumitem := `
	INSERT OR REPLACE INTO items(
		IP,
		RX,
		TX,
		TIME
	) values(
		?, 
		(SELECT RX FROM items WHERE IP=? AND TIME=?)+?, 
		(SELECT TX FROM items WHERE IP=? AND TIME=?)+?, 
		?)
	`
	stmt, err := db.Prepare(sqlAdditem)
	nstmt, err := db.Prepare(sqlSumitem)
	defer stmt.Close()
	defer nstmt.Close()

	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		if ipExists(db, item.ip) {
			_, err := nstmt.Exec(item.ip, item.ip, time.Now().Format(captureMask), item.rx, item.ip, time.Now().Format(captureMask), item.tx, time.Now().Format(captureMask))
			if err != nil {
				log.Fatal(err)
			}

		} else {
			_, err := stmt.Exec(item.ip, item.rx, item.tx, time.Now().Format(captureMask))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func readItem(db *sql.DB, date string) map[string]datausage {
	// read and return data from database
	sqlReadall := `
	SELECT IP, RX, TX, TIME FROM items
	WHERE TIME = ?
	ORDER BY TIME DESC
	`
	stmt, err := db.Prepare(sqlReadall)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query(date)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	result := make(map[string]datausage)
	for rows.Next() {
		item := datausage{}
		if err := rows.Scan(&item.ip, &item.rx, &item.tx, &item.tstamp); err != nil {
			log.Fatal(err)
		}
		result[item.ip] = item
	}
	return result

}

func ipExists(db *sql.DB, qur string) bool {
	sqlStmt := `SELECT IP FROM items WHERE IP = ? and TIME = ?`
	err := db.QueryRow(sqlStmt, qur, time.Now().Format(captureMask)).Scan(&qur)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}
		return false
	}
	return true
}
