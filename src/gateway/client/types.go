package client

type AuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthenticationResponse struct {
	Error string `json:"error,omitempty"`
	Token string `json:"token,omitempty"`
}

type ActiveUsersResponse struct {
	ActiveUsers []string `json:"activeUsers"`
}
