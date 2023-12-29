package auth

//go:generate sh -c "mockgen -package=auth -destination=auth_mock.go . TokenValidator,TokenIssuer"

type Principal struct {
	ID       int    `json:"id"`
	UserName string `json:"userName"`
}

type TokenValidator interface {
	Validate(token string) (Principal, error)
}

type TokenIssuer interface {
	Issue(principal Principal) (string, error)
}
