package auth


type AuthRequest struct {
  UserName string
  Password string
}


type AuthResponse struct {
  Jwt string
  UserAttrs *UserAttrs
}








