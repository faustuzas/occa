package client

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
	ActiveUsers []string `json:"activeUsers"`
}
