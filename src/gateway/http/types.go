package http

import (
	"github.com/faustuzas/occa/src/gateway/services"
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
