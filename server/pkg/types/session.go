package types

import (
	"time"
)

// SessionStatus :
type SessionStatus string

const (
	// SessionValid :
	SessionValid = "Valid"
	// SessionInvalid :
	SessionInvalid = "Invalid"
	// SessionStale :
	SessionStale = "Stale"
)

// Session :
type Session struct {
	ID uint64
	*User
	JWT string
}

// NewSession :
func NewSession(id uint64, jwt string, userAttrs *UserAttrs) *Session {
	return &Session{
		ID:   id,
		User: userAttrs.User,
		JWT:  jwt,
	}
}

// SessionAttrs : not used yet
type SessionAttrs struct {
	ID          uint64
	Session     *Session
	Created     time.Time
	LastConnect time.Time
	Destroyed   time.Time
	Status      SessionStatus
}

// NewSessionAttrs :
func NewSessionAttrs(session *Session) *SessionAttrs {
	// TODO: session ID
	return &SessionAttrs{
		ID:          session.ID,
		Session:     session,
		Created:     time.Now(),
		LastConnect: time.Now(),
		Status:      SessionValid,
	}
}
