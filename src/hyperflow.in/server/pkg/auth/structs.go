package auth

import(
  "time"
)

type UserStatus string

const (
  ValidUser = "Valid"
  InvalidUser = "Invalid"
)

type UserType string
const (
  GuestUser = "Guest"
  StandardUser = "Standard"

)

type SessionStatus string

const (
  SessionValid = "Valid"
  SessionInvalid = "Invalid"
  SessionStale = "Stale"
)

type User struct {
  Name string
}

type UserAttrs struct  {
  *User
  Email string
  PasswordHash string
  Status UserStatus 
  Type UserType
}

type Session struct {
  *User 
  Status SessionStatus
}

// not used yet 
type SessionAttrs struct {
  Id string
  session Session
  created time.Time
  lastConnect time.Time
  destroyed time.Time
}

