package auth

var (
	noopPrincipal = Principal{
		ID:       123,
		UserName: "The_User",
	}
)

var _ TokenIssuer = noopAuth{}
var _ TokenValidator = noopAuth{}

type noopAuth struct {
}

func (a noopAuth) Issue(_ Principal) (string, error) {
	return "token", nil
}

func (a noopAuth) Validate(_ string) (Principal, error) {
	return noopPrincipal, nil
}
