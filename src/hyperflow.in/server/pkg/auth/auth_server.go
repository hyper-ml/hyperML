package auth

import(
  "fmt"

  db_pkg "hyperflow.in/server/pkg/db"
)


type AuthServer interface {
  CreateUser(name, email, password string) (*UserAttrs, error)
  CreateTypedUser(usertype, name, email, plainpass string) (*UserAttrs, error)
  CreateSession(name, txtPassword string) (jwtToken string, userAttrs *UserAttrs, fnError error) 
}


type authServer struct { 
  qs *queryServer
}

func NewAuthServer(db db_pkg.DatabaseContext) AuthServer {
  return &authServer {
    qs: newQueryServer(db),
  }
}

func (a *authServer) CreateUser(name, email, txtPassword string) (*UserAttrs, error) {
  return a.CreateTypedUser(StandardUser, name, email, txtPassword)
}

func (a *authServer) CreateTypedUser(usertype, name, email, plainpass string) (*UserAttrs, error) {
  user_type := UserType(usertype)  

  if user_type == "" {
    user_type = StandardUser
  }

  if err := validatePassword(user_type, plainpass); err != nil {
    return nil, err
  }
  hash, err := hashAndSalt([]byte(plainpass))
  if err != nil {
    return nil, err
  }

  user_attrs := &UserAttrs{
    User: &User{
      Name: name,
    },
    Type: UserType(usertype),
    Email: email,
    PasswordHash: hash,
  }

  user_attrs = a.validateUser(user_attrs)

  err = a.qs.InsertUserAttrs(name, user_attrs)
  if err != nil {
    return nil, err
  }
 
  return a.qs.GetUserAttrs(name)
}

func (a *authServer) validateUser(user_attrs *UserAttrs) *UserAttrs {
  user_attrs.Status = ValidUser 
  return user_attrs
}


func (a *authServer) CreateSession(name, txtPassword string) (jwtToken string, userAttrs *UserAttrs, fnError error) {

  user_attrs, err := a.qs.GetUserAttrs(name) 
  
  if err != nil {
    fmt.Println("[authServer.CreateSession] User get error: ", err)  
    if IsNoDataFoundErr(err) {
      fnError = fmt.Errorf("User does not exist")
      return 
    }

    fnError = err
    return 
  }

  switch{
  case user_attrs == nil:
    fnError = fmt.Errorf("Invalid user")
    return 
  
  case user_attrs.Type == StandardUser &&
       user_attrs.Status == ValidUser:
    
    if !comparePasswords(user_attrs.PasswordHash, []byte(txtPassword)) {
      fnError = fmt.Errorf("Invalid password")
      return 
    } 
    return GenerateToken(user_attrs.User), user_attrs, nil
  
  case user_attrs.Type == GuestUser &&
       user_attrs.Status == ValidUser:
    return GenerateToken(user_attrs.User), user_attrs, nil
  }
  fnError = fmt.Errorf("Invalid user status to create session")
  return 
}



