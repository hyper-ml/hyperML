package types

// UserStatus : User status message
type UserStatus string

// UserType : Standard, Guest or Admin
type UserType string

const (
	// ValidUser : lookup key val
	ValidUser = "Valid"
	// InvalidUser : lookup key val
	InvalidUser = "Invalid"

	// GuestUser : lookup key val
	GuestUser UserType = "Guest"

	// StandardUser : lookup key val
	StandardUser UserType = "Standard"

	// AdminUser : user type
	AdminUser UserType = "Admin"
)

// User : User Object
type User struct {
	Name string
}

// UserAttrs : User attrbutes table
type UserAttrs struct {
	*User
	Email        string
	PasswordHash string
	Status       UserStatus
	Type         UserType
	APIKey       string
}

// UserPOD : holds user pods
type UserPOD struct {
	POD
	User *User
}

// UserPersistentDisk : User specific disk / pv details
type UserPersistentDisk struct {
	User *User
	Disk *PersistentDisk
}
