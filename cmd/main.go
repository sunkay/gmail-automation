package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sunkay11/gmail-automation/internal/config"
	"github.com/sunkay11/gmail-automation/internal/db"
	"github.com/sunkay11/gmail-automation/internal/gmailapi"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("storeInbox --numEmails <number of emails to store> storeDeleted getStored")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Extract the command from os.Args
	command := os.Args[1]

	// Create a new flag set for the command
	cmdFlags := flag.NewFlagSet("command", flag.ExitOnError)

	// Add the flag definition here
	numEmails := cmdFlags.Int("numEmails", 50, "Number of emails to store")

	// Parse the flags
	err = cmdFlags.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	emailDB := db.NewSQLiteDB(cfg.DB.Path)
	// Create a new GmailClient instance
	gmailClient := gmailapi.NewGmailClient(emailDB)

	switch command {
	case "storeInbox":
		err := gmailClient.GetInboxEmailsAndStore(*numEmails)
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
		emails, err := emailDB.GetEmails("emails")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		for i, email := range emails {
			fmt.Printf("[%d], [%s], [%s], [%s]\n", i, email.From, email.Subject, email.SentDate)
		}
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}
