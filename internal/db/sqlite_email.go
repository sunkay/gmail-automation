package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	DB *sql.DB
}

func NewSQLiteDB(filename string) *SQLiteDB {
	db, err := sql.Open("sqlite3", filename)
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
		created_at DATETIME,
		UNIQUE(subject, "from", "to", "sentDate")
	);`

	_, err := s.DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// create the deleted_emails table
	deletedQuery := `CREATE TABLE IF NOT EXISTS deleted_emails (
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
		created_at DATETIME,
		UNIQUE(subject, "from", "to", "sentDate")
	);`

	_, err = s.DB.Exec(deletedQuery)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

}

func (s *SQLiteDB) InsertEmail(email *Email) (int64, error) {
	query := `INSERT OR REPLACE INTO emails 
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

// implementation of batch InsertEmails
func (s *SQLiteDB) InsertEmails(emails []Email) (int64, error) {
	baseQuery := `INSERT OR REPLACE INTO emails 
                (subject, body, "from", "to", "Cc", "Bcc", "sentDate", "sender", "read", "deleted", "labels", created_at)
                VALUES `
	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, email := range emails {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))")
		valueArgs = append(valueArgs, email.Subject, email.Body, email.From, email.To, email.Cc, email.Bcc, email.SentDate, email.Sender, email.Read, email.Deleted, email.Labels)
	}

	query := baseQuery + strings.Join(valueStrings, ",")

	// Start a transaction
	tx, err := s.DB.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return 0, err
	}

	result, err := tx.Exec(query, valueArgs...)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		tx.Rollback()
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		tx.Rollback()
		return 0, err
	}
	log.Printf("Rows affected: %d", rowsAffected)

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		tx.Rollback()
		return 0, err
	}

	return rowsAffected, nil
}

func (s *SQLiteDB) InsertDeletedEmails(emails []Email) (int64, error) {
	baseQuery := `INSERT OR REPLACE INTO deleted_emails 
                (subject, body, "from", "to", "Cc", "Bcc", "sentDate", "sender", "read", "deleted", "labels", created_at)
                VALUES `
	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, email := range emails {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))")
		valueArgs = append(valueArgs, email.Subject, email.Body, email.From, email.To, email.Cc, email.Bcc, email.SentDate, email.Sender, email.Read, email.Deleted, email.Labels)
	}

	query := baseQuery + strings.Join(valueStrings, ",")

	// Start a transaction
	tx, err := s.DB.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return 0, err
	}

	result, err := tx.Exec(query, valueArgs...)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		tx.Rollback()
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		tx.Rollback()
		return 0, err
	}

	return rowsAffected, nil
}

func (s *SQLiteDB) GetEmails(tableName string) ([]Email, error) {
	query := fmt.Sprintf(`SELECT id, 
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
	FROM %s ORDER BY id DESC LIMIT 50`, tableName)

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

func (s *SQLiteDB) GetEmail(tableName string, subject string, from string, to string, sentDate string) (Email, error) {
	query := `SELECT id, 
				"subject", 
				"from", 
				"to", 
				"sentDate",
				"labels"
				FROM emails WHERE subject = $1 AND "from" = $2 AND "to" = $3 AND "sentDate" = $4 ORDER BY created_at DESC`

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
