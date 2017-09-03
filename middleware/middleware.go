package middleware

import "github.com/itchenyi/akita"

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(c akita.Context) bool
)

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(akita.Context) bool {
	return false
}
