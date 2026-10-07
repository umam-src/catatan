package ai

import "context"

// Role adalah peran pesan yang dikirim ke penyedia model.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Response struct {
	Text  string
	Model string
}

type Provider interface {
	Generate(context.Context, Request) (Response, error)
}
