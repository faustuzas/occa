package auth

import "context"

type principalKey int

var key principalKey

func PrincipalFromContext(ctx context.Context) Principal {
	return ctx.Value(key).(Principal)
}

func ContextWithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, key, principal)
}
