package credentials

import (
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GetGmailCredentials() (*oauth2.Config, error) {
	b, err := ioutil.ReadFile("./client_secret.json")
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/gmail.readonly")
	if err != nil {
		return nil, err
	}

	return config, nil
}
