package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type datausage struct {
	ip     string
	tx, rx uint
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
		IP TEXT NOT NULL PRIMARY KEY,
		RX TEXT,
		TX TEXT,
		TIME DATETIME
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
	) values(?, ?, ?, CURRENT_TIMESTAMP)
	`

	sqlSumitem := `
	INSERT OR REPLACE INTO items(
		IP,
		RX,
		TX,
		TIME
	) values(
		?, 
		(SELECT RX FROM items WHERE IP=?)+?, 
		(SELECT TX FROM items WHERE IP=?)+?, 
		CURRENT_TIMESTAMP)
	`
	stmt, err := db.Prepare(sqlAdditem)
	nstmt, err := db.Prepare(sqlSumitem)

	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		if ipExists(db, item.ip) {
			_, err := nstmt.Exec(item.ip, item.ip, item.rx, item.ip, item.tx)
			if err != nil {
				log.Fatal(err)
			}
			defer nstmt.Close()
		} else {
			_, err := stmt.Exec(item.ip, item.rx, item.tx)
			if err != nil {
				log.Fatal(err)
			}
			defer stmt.Close()
		}
	}
}

func readItem(db *sql.DB) map[string]datausage {
	// read and return data from database
	sqlReadall := `
	SELECT IP, RX, TX FROM items
	ORDER BY datetime(TIME) DESC
	`
	rows, err := db.Query(sqlReadall)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	result := make(map[string]datausage)
	for rows.Next() {
		item := datausage{}
		if err := rows.Scan(&item.ip, &item.rx, &item.tx); err != nil {
			log.Fatal(err)
		}
		result[item.ip] = item
	}
	return result

}

func ipExists(db *sql.DB, qur string) bool {
	sqlStmt := `SELECT IP FROM items WHERE IP = ?`
	err := db.QueryRow(sqlStmt, qur).Scan(&qur)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}
		return false
	}
	return true
}
