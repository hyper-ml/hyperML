package api_client

import (
  "fmt"
  "io/ioutil"
  "bytes"
  "encoding/json"
  "hyperflow.in/client/rest_client" 
  auth_server "hyperflow.in/server/pkg/auth"

)


func (c *ApiClient) BasicAuth(name, password string) (jwt string, userAttrs *auth_server.UserAttrs, fnError error) {

  client, _   := rest_client.New(c.serverAddr, c.config.AuthUriPath)

  request := client.Verb("POST", c.jwt) 
  auth_req := &auth_server.AuthRequest{
    UserName: name,
    Password: password,
  }

  json_msg, _ := json.Marshal(&auth_req) 
  _ = request.SetBodyReader(ioutil.NopCloser(bytes.NewReader(json_msg)))

  response := request.Do()
  api_resp, err := response.Raw()

  if err != nil {
    fnError = fmt.Errorf("[BasicAuth] Failed to authenticate: %s", err)
    return 
  }

  auth_response := auth_server.AuthResponse{}
  fnError = json.Unmarshal(api_resp, &auth_response)

  if fnError == nil {
    jwt = auth_response.Jwt
    userAttrs = auth_response.UserAttrs
  }

  return
}