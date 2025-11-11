package user

import "context"

// ctxKey is a type-safe context key using empty struct
// WHY: struct{} takes zero bytes, better than string
// HOW: Define unique type to avoid collisions
type ctxKey struct{}

// WithCtx adds User to context
// WHY: Pass authenticated user through middleware chain
// HOW: Uses type-safe key to avoid collisions with other context values
func WithCtx(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, ctxKey{}, user)
}

// FromCtx retrieves User from context
// WHY: Access authenticated user in handlers
// HOW: Returns (user, true) if found, (nil, false) if not authenticated
//
// Usage:
//   user, ok := user.FromCtx(r.Context())
//   if !ok {
//       // User not authenticated
//       return
//   }
//   // Use user.ID, user.Name, etc.
func FromCtx(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(ctxKey{}).(*User)
	return user, ok
}
