package sqlite

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func ConnectDB() (*sql.DB, error) {
	dbPath := "./pkg/models/sqlite/forum.db"

	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		log.Println("Creating database...")
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		file.Close()
		log.Println("forum.db created")
	} else {
		log.Println("Using existing database")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Read the schema.sql file
	schema, err := os.ReadFile("./pkg/models/sqlite/schema.sql")
	if err != nil {
		return nil, err
	}

	// Execute the schema.sql content as SQL statements
	_, err = db.Exec(string(schema))
	if err != nil {
		return nil, err
	}
	return db, nil
}
