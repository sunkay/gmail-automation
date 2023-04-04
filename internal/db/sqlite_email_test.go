package db

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestInsertEmails(t *testing.T) {
	// Create a new SQLiteDB instance
	db := NewSQLiteDB("./test_emails.sqlite")
	// Clean up the test database file
	defer cleanupTestDB("./test_emails.sqlite")

	// Test data for the emails
	emails := []Email{
		{
			Subject:  "Test Email 1",
			Body:     "Hello, this is a test email 1.",
			From:     "test1@example.com",
			To:       "recipient1@example.com",
			Cc:       "",
			Bcc:      "",
			SentDate: time.Now().Format(time.RFC1123Z),
		},
		{
			Subject:  "Test Email 2",
			Body:     "Hello, this is a test email 2.",
			From:     "test2@example.com",
			To:       "recipient2@example.com",
			Cc:       "",
			Bcc:      "",
			SentDate: time.Now().Add(-1 * time.Hour).Format(time.RFC1123Z),
		},
	}

	// Call the InsertEmails function
	rowsAffected, err := db.InsertEmails(emails)
	if err != nil {
		t.Fatalf("InsertEmails failed: %v", err)
	}

	// Check if the correct number of emails was inserted
	if int(rowsAffected) != len(emails) {
		t.Fatalf("Expected %d inserted emails, got %d", len(emails), rowsAffected)
	}

	// Retrieve emails from the database and check if they match the test data
	storedEmails, err := db.GetEmails("emails")
	if err != nil {
		t.Fatalf("GetEmails failed: %v", err)
	}

	// Check if the number of retrieved emails matches the number of inserted emails
	if len(storedEmails) != len(emails) {
		t.Fatalf("Expected %d stored emails, got %d", len(emails), len(storedEmails))
	}

	// Compare the stored emails with the test data
	for i, email := range emails {
		storedEmail := storedEmails[len(storedEmails)-1-i] // Get the email in reverse order

		if email.Subject != storedEmail.Subject || email.Body != storedEmail.Body || email.From != storedEmail.From || email.To != storedEmail.To || email.Cc != storedEmail.Cc || email.Bcc != storedEmail.Bcc || email.SentDate != storedEmail.SentDate {
			t.Errorf("Email mismatch. Expected: %+v, Got: %+v", email, storedEmail)
		}
	}
}

// This is to test INSERT AND REPLACE functionality
// Insert 3 emails but only 2 are unique
func TestInsertDuplicateEmails(t *testing.T) {
	// Create a new SQLiteDB instance
	db := NewSQLiteDB("./test_emails.sqlite")
	// Clean up the test database file
	defer cleanupTestDB("./test_emails.sqlite")

	// Test data for the emails
	emails := []Email{
		{
			Subject:  "Test Email 1",
			Body:     "Hello, this is a test email 1.",
			From:     "test1@example.com",
			To:       "recipient1@example.com",
			Cc:       "",
			Bcc:      "",
			Labels:   "X, Y, Z",
			SentDate: time.Now().Format(time.RFC1123Z),
		},
		{
			Subject:  "Test Email 2",
			Body:     "Hello, this is a test email 2.",
			From:     "test2@example.com",
			To:       "recipient2@example.com",
			Cc:       "",
			Bcc:      "",
			Labels:   "X, Y, Z",
			SentDate: time.Now().Add(-1 * time.Hour).Format(time.RFC1123Z),
		},
		{
			Subject:  "Test Email 2",
			Body:     "Hello, this is a test email 2.",
			From:     "test2@example.com",
			To:       "recipient2@example.com",
			Cc:       "",
			Bcc:      "",
			Labels:   "A, B, C",
			SentDate: time.Now().Add(-1 * time.Hour).Format(time.RFC1123Z),
		},
	}

	// Call the InsertEmails function
	rowsAffected, err := db.InsertEmails(emails)
	if err != nil {
		t.Fatalf("InsertEmails failed: %v", err)
	}

	// Check if the correct number of emails was inserted
	if int(rowsAffected) != len(emails) {
		t.Fatalf("Expected %d inserted emails, got %d", len(emails)-1, rowsAffected)
	}

	// Retrieve emails from the database and check if they match the test data
	storedEmails, err := db.GetEmails("emails")
	if err != nil {
		t.Fatalf("GetEmails failed: %v", err)
	}

	// Check if the number of retrieved emails matches the number of inserted emails
	if len(storedEmails) != len(emails)-1 {
		t.Fatalf("Expected %d stored emails, got %d", len(emails)-1, len(storedEmails))
	}

	// Compare the stored emails with the test data
	for i, email := range emails {
		// break at len(emails)-1 because the last email is a duplicate
		if i == len(emails)-1 {
			break
		}

		storedEmail := storedEmails[len(storedEmails)-1-i] // Get the email in reverse order
		if email.Subject != storedEmail.Subject || email.Body != storedEmail.Body || email.From != storedEmail.From || email.To != storedEmail.To || email.Cc != storedEmail.Cc || email.Bcc != storedEmail.Bcc || email.SentDate != storedEmail.SentDate {
			t.Errorf("Email mismatch. Expected: %+v, Got: %+v", email, storedEmail)
		}
	}

	// Lets insert the duplicate email again
	emails = []Email{
		{
			Subject:  "Test Email 1",
			Body:     "Hello, this is a test email 1.",
			From:     "test1@example.com",
			To:       "recipient1@example.com",
			Cc:       "",
			Bcc:      "",
			Labels:   "D, E, F",
			SentDate: time.Now().Format(time.RFC1123Z),
		},
		{
			Subject:  "Test Email 2",
			Body:     "Hello, this is a test email 2.",
			From:     "test2@example.com",
			To:       "recipient2@example.com",
			Cc:       "",
			Bcc:      "",
			Labels:   "X, Y, Z",
			SentDate: time.Now().Add(-1 * time.Hour).Format(time.RFC1123Z),
		},
		{
			Subject:  "Test Email 5",
			Body:     "Hello, this is a test email 5.",
			From:     "test5@example.com",
			To:       "recipient5@example.com",
			Cc:       "",
			Bcc:      "",
			Labels:   "X, Y, Z",
			SentDate: time.Now().Add(-1 * time.Hour).Format(time.RFC1123Z),
		},
	}
	rowsAffected, err = db.InsertEmails(emails)
	if err != nil {
		t.Fatalf("InsertEmails failed: %v", err)
	}

	// Check if the correct number of emails was inserted
	if int(rowsAffected) != len(emails) {
		t.Fatalf("Expected %d inserted emails, got %d", len(emails)-1, rowsAffected)
	}

	// Retrieve emails from the database and check if they match the test data
	storedEmails, err = db.GetEmails("emails")
	if err != nil {
		t.Fatalf("GetEmails failed: %v", err)
	}

	// Compare the stored emails with the test data
	for i, email := range emails {

		storedEmail := storedEmails[len(storedEmails)-1-i] // Get the email in reverse order

		if email.Subject != storedEmail.Subject || email.Body != storedEmail.Body || email.From != storedEmail.From || email.To != storedEmail.To || email.Cc != storedEmail.Cc || email.Bcc != storedEmail.Bcc || email.SentDate != storedEmail.SentDate {
			t.Errorf("Email mismatch. Expected: %+v, Got: %+v", email, storedEmail)
		}
	}

}

func TestGetEmails(t *testing.T) {
	// Create a new SQLiteDB instance
	db := NewSQLiteDB("./test_emails.sqlite")
	// Clean up the test database file
	defer cleanupTestDB("./test_emails.sqlite")

	// Test data for the emails
	emails := []Email{
		{
			Subject:  "Test Email 1",
			Body:     "Hello, this is a test email 1.",
			From:     "test1@example.com",
			To:       "recipient1@example.com",
			Cc:       "",
			Bcc:      "",
			SentDate: time.Now().Format(time.RFC1123Z),
		},
		{
			Subject:  "Test Email 2",
			Body:     "Hello, this is a test email 2.",
			From:     "test2@example.com",
			To:       "recipient2@example.com",
			Cc:       "",
			Bcc:      "",
			SentDate: time.Now().Add(-1 * time.Hour).Format(time.RFC1123Z),
		},
	}

	// Call the InsertEmails function
	_, err := db.InsertEmails(emails)
	if err != nil {
		t.Fatalf("InsertEmails failed: %v", err)
	}

	// Get the emails from the database
	resultEmails, err := db.GetEmails("emails")
	if err != nil {
		t.Errorf("Error getting emails: %v", err)
	}

	// Verify that the correct number of emails were retrieved
	if len(resultEmails) != len(emails) {
		t.Errorf("Expected %d emails, but got %d", len(emails), len(resultEmails))
	}

	// Verify that the retrieved emails are the same as the inserted emails
	expectedEmails := make(map[string]Email)
	for _, email := range emails {
		key := fmt.Sprintf("%s:%s:%s:%s", email.Subject, email.From, email.To, email.SentDate)
		expectedEmails[key] = email
	}

	for _, email := range resultEmails {
		key := fmt.Sprintf("%s:%s:%s:%s", email.Subject, email.From, email.To, email.SentDate)
		_, ok := expectedEmails[key]
		if !ok {
			t.Errorf("Unexpected email %+v", email)
			continue
		}
	}

}

func TestInsertEmailAndCheckForDuplicates(t *testing.T) {
	// Create a new SQLiteDB instance
	testDB := NewSQLiteDB("./test_emails.sqlite")
	// Clean up the test database file
	defer cleanupTestDB("./test_emails.sqlite")

	// create a new email
	email := &Email{
		Subject:   "Test Subject",
		Body:      "Test Body",
		From:      "test@example.com",
		To:        "recipient@example.com",
		Cc:        "",
		Bcc:       "",
		SentDate:  "2023-04-03T08:00:00Z",
		Sender:    "",
		Read:      false,
		Deleted:   false,
		Labels:    "",
		CreatedAt: "",
	}

	// insert the email
	id, err := testDB.InsertEmail(email)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// check that the email was inserted with the correct id
	if id != 1 {
		t.Errorf("Expected id 1, but got %d", id)
	}

	// insert the same email again, which should result in an error
	_, err = testDB.InsertEmail(email)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Get the emails from the database
	resultEmails, _ := testDB.GetEmails("emails")
	if len(resultEmails) != 1 {
		t.Errorf("Expected 1 but got: %d", len(resultEmails))
	}
}

func cleanupTestDB(filename string) {
	err := os.Remove(filename)
	if err != nil {
		log.Fatalf("Failed to clean up test database file: %v", err)
	}
}

func TestInsertDeletedEmails(t *testing.T) {
	// Create a new SQLiteDB instance
	db := NewSQLiteDB("./test_emails.sqlite")
	// Clean up the test database file
	defer cleanupTestDB("./test_emails.sqlite")

	// Test data for the emails
	emails := []Email{
		{
			Subject:  "Test Email 1",
			Body:     "Hello, this is a test email 1.",
			From:     "test1@example.com",
			To:       "recipient1@example.com",
			Cc:       "",
			Bcc:      "",
			SentDate: time.Now().Format(time.RFC1123Z),
		},
		{
			Subject:  "Test Email 2",
			Body:     "Hello, this is a test email 2.",
			From:     "test2@example.com",
			To:       "recipient2@example.com",
			Cc:       "",
			Bcc:      "",
			SentDate: time.Now().Add(-1 * time.Hour).Format(time.RFC1123Z),
		},
	}

	// Call the InsertEmails function
	rowsAffected, err := db.InsertDeletedEmails(emails)
	if err != nil {
		t.Fatalf("InsertDeletedEmails failed: %v", err)
	}

	// Check if the correct number of emails was inserted
	if int(rowsAffected) != len(emails) {
		t.Fatalf("Expected %d inserted emails, got %d", len(emails), rowsAffected)
	}

	// Retrieve emails from the database and check if they match the test data
	storedEmails, err := db.GetEmails("deleted_emails")
	if err != nil {
		t.Fatalf("GetEmails failed: %v", err)
	}

	// Check if the number of retrieved emails matches the number of inserted emails
	if len(storedEmails) != len(emails) {
		t.Fatalf("Expected %d stored emails, got %d", len(emails), len(storedEmails))
	}

	// Compare the stored emails with the test data
	for i, email := range emails {
		storedEmail := storedEmails[len(storedEmails)-1-i] // Get the email in reverse order

		if email.Subject != storedEmail.Subject || email.Body != storedEmail.Body || email.From != storedEmail.From || email.To != storedEmail.To || email.Cc != storedEmail.Cc || email.Bcc != storedEmail.Bcc || email.SentDate != storedEmail.SentDate {
			t.Errorf("Email mismatch. Expected: %+v, Got: %+v", email, storedEmail)
		}
	}
}

// TestGetEmail given a subject, from, to, and sent date, the GetEmail function should return the corresponding email
func TestGetEmail(t *testing.T) {
	// Create a new SQLiteDB instance
	db := NewSQLiteDB("./test_emails.sqlite")
	// Clean up the test database file
	defer cleanupTestDB("./test_emails.sqlite")

	// Insert a test email into the database
	testEmail := Email{
		Subject:  "Test Email",
		From:     "test@example.com",
		To:       "recipient@example.com",
		SentDate: "2023-04-03T08:58:29-04:00",
		Labels:   "inbox",
	}
	_, err := db.InsertEmail(&testEmail)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Test that the inserted email can be retrieved
	resultEmail, err := db.GetEmail("emails", testEmail.Subject, testEmail.From, testEmail.To, testEmail.SentDate)
	if err != nil {
		t.Fatalf("Failed to retrieve test email: %v", err)
	}

	// Verify that the retrieved email matches the inserted email
	if resultEmail.Subject != testEmail.Subject || resultEmail.From != testEmail.From || resultEmail.To != testEmail.To || resultEmail.SentDate != testEmail.SentDate {
		t.Errorf("Expected email %+v, but got %+v", testEmail, resultEmail)
	}
}
