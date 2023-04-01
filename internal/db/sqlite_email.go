package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	DB *sql.DB
}

func NewSQLiteDB() *SQLiteDB {
	db, err := sql.Open("sqlite3", "./emails.sqlite")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	sqliteDB := &SQLiteDB{DB: db}
	sqliteDB.createTable()

	return sqliteDB
}

func (s *SQLiteDB) createTable() {
	query := `CREATE TABLE IF NOT EXISTS emails (
		id INTEGER PRIMARY KEY,
		"subject" TEXT,
		"body" TEXT,
		"from" TEXT,
		"to"	TEXT,
		"Cc" TEXT,
		"Bcc" TEXT,
		"sentDate" TEXT,
		"sender" TEXT,
		created_at DATETIME
	);`

	_, err := s.DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func (s *SQLiteDB) InsertEmail(email *Email) error {
	query := `INSERT INTO emails 
				("subject", "body", "from", "to", "Cc", "Bcc", "sentDate", "sender", created_at) 
			  	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, datetime('now')) 
				RETURNING id`
	var id int64
	err := s.DB.QueryRow(query,
		email.Subject,
		email.Body,
		email.From,
		email.To,
		email.Cc,
		email.Bcc,
		email.SentDate,
		email.Sender,
	).Scan(&id)
	if err != nil {
		return err
	}

	email.Id = id
	return nil
}

func (s *SQLiteDB) GetEmails() ([]Email, error) {
	query := `SELECT id, 
				"subject", 
				"body",
				"from", 
				"to", 
				"Cc", 
				"Bcc", 
				"sentDate", 
				"sender", 
				created_at			
				FROM emails ORDER BY id DESC LIMIT 50`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emails := []Email{}
	for rows.Next() {
		var email Email
		err = rows.Scan(&email.Id,
			&email.Subject,
			&email.Body,
			&email.From,
			&email.To,
			&email.Cc,
			&email.Bcc,
			&email.SentDate,
			&email.Sender,
			&email.CreatedAt)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	return emails, nil
}
