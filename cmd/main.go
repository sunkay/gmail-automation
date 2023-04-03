package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sunkay11/gmail-automation/internal/db"
	"github.com/sunkay11/gmail-automation/internal/gmailapi"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gmail-automation top5 storeDeleted getStored")
		os.Exit(1)
	}

	command := os.Args[1]

	emailDB := db.NewSQLiteDB()

	// Create a new GmailClient instance
	gmailClient := gmailapi.NewGmailClient(emailDB)

	switch command {
	case "storeInbox":
		err := gmailClient.GetInboxEmailsAndStore(1)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "storeDeleted":
		err := gmailClient.GetDeletedEmailsAndStore(1)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "getStored":
		emails, err := emailDB.GetEmails()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		for i, email := range emails {
			fmt.Printf("[%d], [%s], [%s], [%s]\n", i, email.From, email.Subject, email.SentDate)
		}
	default:
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}
}
