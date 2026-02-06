package main

import (
	"excel_core/exporter"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func main() {
	// 1. Init sqlite in-memory
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Create table
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		email TEXT,
		age INTEGER,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);
	`
	db.MustExec(schema)

	// 3. Insert sample data
	tx := db.MustBegin()
	for i := 1; i <= 5000; i++ {
		tx.MustExec(
			`INSERT INTO users (username, email, age, created_at, updated_at, deleted_at)
			 VALUES (?, ?, ?, datetime('now'), datetime('now'), datetime('now'))`,
			fmt.Sprintf("user_%d", i),
			fmt.Sprintf("user_%d@example.com", i),
			20+i%10,
		)
	}
	tx.Commit()

	// 4. Query rows
	rows, err := db.Queryx(`SELECT id, username, email, age, created_at, updated_at, deleted_at FROM users`)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Call Excel stream export
	filePath, err := exporter.CreateExcelStream(rows, "user_report", "BÁO CÁO QUẢN LÝ NGƯỜI DÙNG")

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Excel generated successfully:", filePath)
}
