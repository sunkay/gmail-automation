package openai

type GPT interface {
	ClassifyEmail(prompt string) (string, error)
}
