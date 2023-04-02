package gmailapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sunkay11/gmail-automation/internal/credentials"
	"github.com/sunkay11/gmail-automation/internal/db"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailClient struct {
	emailDB db.EmailDB
}

func NewGmailClient(emailDB db.EmailDB) *GmailClient {
	return &GmailClient{emailDB: emailDB}
}

func (gc *GmailClient) GetTop5Emails() error {
	return getTop5Emails(gc.emailDB)
}

func (gc *GmailClient) GetDeletedEmails(daysAgo int) error {
	return getDeletedEmails(gc.emailDB, daysAgo)
}

func getTop5Emails(database db.EmailDB) error {
	config, err := credentials.GetGmailCredentials()
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return err
	}
	client := getClient(config)

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	query := "in:inbox is:unread OR is:read OR is:Deleted"
	messages, err := srv.Users.Messages.List(user).MaxResults(50).Q(query).Do()
	if err != nil {
		return err
	}

	log.Println("Total messages:", len(messages.Messages))
	for i, message := range messages.Messages {
		msg, err := srv.Users.Messages.Get(user, message.Id).Fields("labelIds, payload/headers").Do()
		if err != nil {
			log.Printf("Failed to get message: %v", err)
			continue
		}

		// If you need to check specific labels, you can do so using conditional statements.
		read := isLabelPresent(msg.LabelIds, "UNREAD")
		deleted := isLabelPresent(msg.LabelIds, "TRASH")
		labels := strings.Join(msg.LabelIds, ", ")

		headers := make(map[string]string)
		for _, header := range msg.Payload.Headers {
			headers[header.Name] = header.Value
		}

		email := db.Email{
			Subject:  headers["Subject"],
			From:     headers["From"],
			To:       headers["To"],
			Cc:       headers["Cc"],
			Bcc:      headers["Bcc"],
			SentDate: headers["Date"],
			Body:     msg.Snippet,
			Sender:   headers["From"],
			Read:     read,
			Deleted:  deleted,
			Labels:   labels,
		}

		email.Id, err = database.InsertEmail(&email)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate email detected") {
				// Update the read status & labels of the existing email if not present
				// check if email.Id is 0, if so, get the email from the database
				var emailCopy db.Email
				if email.Id == 0 {
					emailCopy, err = database.GetEmail(email.Subject, email.From, email.To, email.SentDate)
					if err != nil {
						log.Printf("[%d] Error getting email from the database: %v", i, err)
						continue
					}
					email.Id = emailCopy.Id
					// continue if labels are already present
					if strings.Contains(emailCopy.Labels, email.Labels) {
						continue
					}
				}
				_ = database.UpdateEmailReadStatus(email.Id, email.Read)
				err = database.UpdateEmailLabels(email.Id, email.Labels)
				if err != nil {
					log.Printf("[%d] Error updating read or label status: %v", email.Id, err)
				} else {
					log.Printf("[%d] Updated read or labels for duplicate email", email.Id)
				}
			} else {
				log.Printf("[%d] Error inserting email into the database: %v", email.Id, err)
			}
			continue
		} else {
			log.Printf("[%d] Inserted email into the database [%s]", i, email.Subject)
		}
	}
	return nil
}

// GetDeletedEmails retrieves all deleted emails from the specified number of days ago.
func getDeletedEmails(db db.EmailDB, daysAgo int) error {
	config, err := credentials.GetGmailCredentials()
	if err != nil {
		return err
	}

	//client := config.Client(context.Background(), token)
	client := getClient(config)

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	q := fmt.Sprintf("in:trash before:%s", time.Now().AddDate(0, 0, -daysAgo).Format("2006/01/02"))
	messages, err := srv.Users.Messages.List(user).Q(q).Fields("messages(id,payload/headers)").Do()
	if err != nil {
		return err
	}

	fmt.Println("Deleted emails:")
	for _, message := range messages.Messages {
		msg, err := srv.Users.Messages.Get(user, message.Id).Format("full").Fields("payload/headers").Do()
		if err != nil {
			return err
		}
		fmt.Println("Subject:", getHeader("Subject", msg.Payload.Headers))
	}

	return nil
}

func getHeader(name string, headers []*gmail.MessagePartHeader) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func getTokenFromFile(filename string) (*oauth2.Token, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	return token, err
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := getTokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		saveToken(tokFile, tok)
		if err != nil {
			log.Fatalf("Unable to save token: %v", err)
		}
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	var token *oauth2.Token

	// Create a channel to receive the authorization code.
	codeChan := make(chan string)

	// Set up a temporary web server to handle the OAuth 2.0 redirect.
	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received code")
		code := r.URL.Query().Get("code")
		codeChan <- code
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You can close this window now."))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("got /")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This is doing nothing for you..."))
	})

	// Start the temporary web server.
	listener, err := net.Listen("tcp", "localhost:4000")
	if err != nil {
		return nil, err
	}

	go http.Serve(listener, nil)
	defer listener.Close()

	// Open the user's web browser to the authorization URL.
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Visit the following URL to authorize the app:")
	fmt.Println(authURL)

	// Wait for the authorization code.
	code := <-codeChan

	// Exchange the authorization code for an access token.
	token, err = config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Helper function to check if a specific label is present in the labelIds list
func isLabelPresent(labelIds []string, label string) bool {
	for _, id := range labelIds {
		if id == label {
			return true
		}
	}
	return false
}
