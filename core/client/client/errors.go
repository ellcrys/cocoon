package client

import (
	"fmt"
)

var (

	// ErrNoUserSession tells us about the current user not having an active session
	ErrNoUserSession = fmt.Errorf("user has no active session")
)
