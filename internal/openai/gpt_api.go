package openai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sunkay11/gmail-automation/internal/db"
)

type GPT3Classifier struct {
	client *openai.Client
	model  string
}

func NewGPT3Classifier(apiKey string, model string) *GPT3Classifier {
	client := openai.NewClient(apiKey)
	return &GPT3Classifier{client, model}
}

func (g *GPT3Classifier) ClassifyEmail(prompt string) (string, error) {
	ctx := context.Background()

	completionReq := openai.CompletionRequest{
		Model:     g.model,
		Prompt:    prompt,
		MaxTokens: 100,
		N:         1,
		Stop:      []string{"\n"},
	}

	completion, err := g.client.CreateCompletion(ctx, completionReq)
	if err != nil {
		return "", fmt.Errorf("error creating completion: %v", err)
	}

	if len(completion.Choices) > 0 {
		answer := completion.Choices[0].Text
		return answer, nil
	}

	return "", fmt.Errorf("no answer found")
}

func (g *GPT3Classifier) GenerateContextualPrompt(email db.Email) string {
	prompt := `I have a list of emails, and I need to determine if they should be placed in the trash folder. Emails can have the following labels: UNREAD, CATEGORY_UPDATES, CATEGORY_PROMOTIONS, CATEGORY_PERSONAL, CATEGORY_SOCIAL, CATEGORY_FORUMS, and others. The trash folder typically contains emails that are not important, spam, or promotional in nature.

Please analyze the following email based on its subject, recipients, sender, and labels, and determine if it should be placed in the trash folder.

Email Details:
Subject: %s
To: %s
From: %s
Labels: %s

Based on the email details, should this email be placed in the trash folder? Provide only the label that you think is most appropriate.
`

	return fmt.Sprintf(
		prompt,
		email.Subject,
		email.To,
		email.From,
		email.Labels,
	)
}
