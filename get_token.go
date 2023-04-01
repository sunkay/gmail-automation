package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/sunkay11/gmail-automation/internal/credentials"
	"golang.org/x/oauth2"
)

func main() {
	config, err := credentials.GetGmailCredentials()
	if err != nil {
		fmt.Println("Error getting credentials:", err)
		return
	}

	// Set up a temporary web server to handle the OAuth 2.0 redirect.
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Visit the following URL to authorize the app:")
	fmt.Println(authURL)

	token, err := getTokenFromWeb(config)
	if err != nil {
		fmt.Println("Error getting token:", err)
		return
	}

	saveTokenToFile("./token.json", token)
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

func getTokenFromWeb_2(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Visit the following URL to authorize the app:")
	fmt.Println(authURL)
	fmt.Println("Enter the authorization code:")

	var authCode string
	_, err := fmt.Scan(&authCode)
	if err != nil {
		return nil, err
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func saveTokenToFile(filename string, token *oauth2.Token) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("Error saving token to file:", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(token)
}
