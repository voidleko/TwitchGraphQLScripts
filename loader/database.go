package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func prepareDatabse() {
	log.Printf("Prepate database")
	// Open the database file
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a table
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(20) NOT NULL PRIMARY KEY,
		viewersCount INTEGER NOT NULL,
		login VARCHAR(20) NOT NULL,
		followersCount INTEGER NOT NULL 
    );

    CREATE TABLE IF NOT EXISTS viewers (
		login VARCHAR(20) NOT NULL,
		streamerId VARCHAR(20) NOT NULL,
		time VARCHAR(20) NOT NULL
    );
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}

	log.Printf("Database and table created successfully")
}
