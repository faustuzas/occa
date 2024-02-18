package http

import (
	"github.com/faustuzas/occa/src/gateway/services"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

type RegistrationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegistrationResponse struct {
	Error string `json:"error,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

type ActiveUsersResponse struct {
	ActiveUsers []services.ActiveUser `json:"activeUsers"`
}

type SelectServerResponse struct {
	Address string `json:"address"`
}

type SendMessageRequest struct {
	RecipientID pkgid.ID `json:"recipientId"`
	Message     string   `json:"message"`
}
