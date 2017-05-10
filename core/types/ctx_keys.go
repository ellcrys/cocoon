package types

// CtxKey represents keys used as context value keys
type CtxKey string

const (

	// CtxAuthToken represents an auth token from a request
	CtxAuthToken CtxKey = "auth_token"

	// CtxSessionID represents the session id from the auth token
	CtxSessionID CtxKey = "session_id"

	// CtxIdentity represents a user identity id
	CtxIdentity CtxKey = "identity"
)
