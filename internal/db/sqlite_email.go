package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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
		"read" BOOLEAN DEFAULT 0,
		"deleted" BOOLEAN DEFAULT 0,
		"labels" TEXT,
		created_at DATETIME
	);`

	_, err := s.DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// Create a unique index on subject, from, and to and sentDate columns
	indexQuery := `CREATE UNIQUE INDEX IF NOT EXISTS idx_subject_from_to_sentdate ON emails (subject, "from", "to", "sentDate");`
	_, err = s.DB.Exec(indexQuery)
	if err != nil {
		log.Fatal("Failed to create unique index:", err)
	}
}

func (s *SQLiteDB) InsertEmail(email *Email) (int64, error) {
	query := `INSERT OR IGNORE INTO emails 
				(subject, body, "from", "to", "Cc", "Bcc", "sentDate", "sender", "read", "deleted", "labels", created_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, datetime('now'))`

	result, err := s.DB.Exec(query, email.Subject, email.Body,
		email.From, email.To, email.Cc,
		email.Bcc, email.SentDate, email.Sender, email.Read, email.Deleted, email.Labels)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if id == 0 {
		return 0, fmt.Errorf("duplicate email detected")
	}

	return id, nil
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
				"read",
				"deleted",
				"labels",
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
			&email.Read,
			&email.Deleted,
			&email.Labels,
			&email.CreatedAt)
		if err != nil {
			return nil, err
		}

		emails = append(emails, email)
	}

	return emails, nil
}

func (s *SQLiteDB) UpdateEmailReadStatus(id int64, read bool) error {
	query := `UPDATE emails SET read = $1 WHERE id = $2`
	_, err := s.DB.Exec(query, read, id)
	return err
}

func (s *SQLiteDB) UpdateEmailLabels(id int64, labels string) error {
	query := `UPDATE emails SET labels = $1 WHERE id = $2`

	_, err := s.DB.Exec(query, labels, id)
	return err
}

func (s *SQLiteDB) GetEmail(subject string, from string, to string, sentDate string) (Email, error) {
	query := `SELECT id, 
				"subject", 
				"from", 
				"to", 
				"sentDate",
				"labels"
				FROM emails WHERE subject = $1 AND "from" = $2 AND "to" = $3 AND "sentDate" = $4`

	var email Email
	err := s.DB.QueryRow(query, subject, from, to, sentDate).Scan(&email.Id,
		&email.Subject,
		&email.From,
		&email.To,
		&email.SentDate,
		&email.Labels)
	if err != nil {
		return email, err
	}

	return email, nil

}

// implementation of batch InsertEmails
func (s *SQLiteDB) InsertEmails(emails []Email) ([]int64, error) {
	query := `INSERT INTO emails (subject, body, "from", "to", "Cc", "Bcc", "sentDate", "sender", "read", "deleted", "labels", created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Start a transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	ids := make([]int64, 0, len(emails))
	for _, email := range emails {
		result, err := stmt.Exec(email.Subject, email.Body, email.From, email.To, email.Cc, email.Bcc, email.SentDate, email.Sender, email.Read, email.Deleted, email.Labels, time.Now())
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ids, nil
}
