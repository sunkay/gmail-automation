package openai

type GPT interface {
	ClassifyText(prompt string) (string, error)
}
