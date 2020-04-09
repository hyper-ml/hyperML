package auth

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/qs"
	"github.com/hyper-ml/hyperml/server/pkg/requests"
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

// Server : Manage Authorization
type Server interface {

	// CLI
	CreateUser(name, email, password string) (*types.UserAttrs, error)
	CreateTypedUser(usertype types.UserType, name, email, plainpass string) (*types.UserAttrs, error)
	DisableUser(username string) error
	EnableUser(username string) error
	ShowUser(username string) (*types.UserAttrs, error)

	// Web
	CreateJWT(name, txtPassword string) (jwtToken string, userAttrs *types.UserAttrs, fnError error)
	CreateAndLoginUser(name, email, password string) (*types.SessionAttrs, *types.UserAttrs, error)
	CreateSession(name, txtPassword string) (*types.SessionAttrs, error)
	SaveSession(jwt string, userAttrs *types.UserAttrs) (*types.SessionAttrs, error)
}

type authServer struct {
	noAuth bool
	sqs    *sessionQueryServer
	udh    *requests.UserDataStore
}

// NewAuthServer : Returns new auth server object
func NewAuthServer(q *qs.QueryServer, noAuth bool) Server {
	return &authServer{
		sqs:    newSessionQueryServer(q),
		udh:    requests.NewUserDataStore(q),
		noAuth: noAuth,
	}
}

func (a *authServer) CreateUser(name, email, txtPassword string) (*types.UserAttrs, error) {
	return a.CreateTypedUser(types.StandardUser, name, email, txtPassword)
}

func (a *authServer) CreateTypedUser(usertype types.UserType, name, email, plainpass string) (*types.UserAttrs, error) {

	if string(usertype) == "" {
		usertype = types.StandardUser
	}

	if err := validatePassword(usertype, plainpass); err != nil {
		return nil, err
	}
	hash, err := hashAndSalt([]byte(plainpass))
	if err != nil {
		return nil, err
	}

	userAttrs := &types.UserAttrs{
		User: &types.User{
			Name: name,
		},
		Type:         types.UserType(usertype),
		Email:        email,
		PasswordHash: hash,
	}

	userAttrs = a.validateUser(userAttrs)

	err = a.udh.InsertUserAttrs(name, userAttrs)
	if err != nil {
		return nil, err
	}

	return a.udh.GetUserAttrs(name)
}

// validateUser :
func (a *authServer) validateUser(userAttrs *types.UserAttrs) *types.UserAttrs {
	userAttrs.Status = types.ValidUser
	return userAttrs
}

// CreateSession :
func (a *authServer) CreateJWT(name, txtPassword string) (jwtToken string, userAttrs *types.UserAttrs, fnError error) {

	userAttrs, err := a.udh.GetUserAttrs(name)

	if err != nil {
		fmt.Println("[authServer.CreateSession] User get error: ", err)
		if IsNoDataFoundErr(err) {
			fnError = fmt.Errorf("User does not exist")
			return
		}

		fnError = err
		return
	}

	switch {
	case userAttrs == nil:
		fnError = fmt.Errorf("Invalid user")
		return

	case userAttrs.Type == types.StandardUser &&
		userAttrs.Status == types.ValidUser:

		if !a.noAuth && !comparePasswords(userAttrs.PasswordHash, []byte(txtPassword)) {
			fnError = fmt.Errorf("Invalid password")
			return
		}
		return GenerateToken(userAttrs.User), userAttrs, nil

	case userAttrs.Type == types.GuestUser &&
		userAttrs.Status == types.ValidUser:
		return GenerateToken(userAttrs.User), userAttrs, nil
	}

	fnError = fmt.Errorf("Invalid user status to create session")
	return
}

func (a *authServer) SaveSession(jwt string, userAttrs *types.UserAttrs) (*types.SessionAttrs, error) {
	id, err := a.sqs.NewSessionID()

	fmt.Println("session ID:", id)

	if err != nil {
		return nil, err
	}

	if id == 0 {
		return nil, fmt.Errorf("Failed to generate session ID")
	}

	session := types.NewSession(id, jwt, userAttrs)

	sessAttrs := types.NewSessionAttrs(session)
	return sessAttrs, nil
}

func (a *authServer) CreateSession(name, txtPassword string) (*types.SessionAttrs, error) {
	jwt, userAttrs, err := a.CreateJWT(name, txtPassword)
	if err != nil {
		return nil, err
	}
	return a.SaveSession(jwt, userAttrs)
}

// CreateAndLoginUser : creates and logs in the new user
func (a *authServer) CreateAndLoginUser(name, email, password string) (*types.SessionAttrs, *types.UserAttrs, error) {
	var jwt string

	userAttrs, err := a.CreateUser(name, email, password)
	if err != nil {
		return nil, nil, err
	}

	jwt, userAttrs, err = a.CreateJWT(name, password)
	if err != nil {
		return nil, userAttrs, err
	}

	sessAttrs, err := a.SaveSession(jwt, userAttrs)
	if err != nil {
		return nil, userAttrs, err
	}

	return sessAttrs, userAttrs, nil
}

// DisableUser : Disable a given user
func (a *authServer) DisableUser(username string) error {
	return fmt.Errorf("unimplemeted")
}

// EnableUser : Enable a given user
func (a *authServer) EnableUser(username string) error {
	return fmt.Errorf("unimplemeted")
}

// ShowUser : Returns user record
func (a *authServer) ShowUser(username string) (*types.UserAttrs, error) {
	return nil, fmt.Errorf("unimplemeted")
}
