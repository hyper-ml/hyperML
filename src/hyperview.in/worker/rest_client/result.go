package rest_client

import ( 
  "fmt"
  "encoding/json"

  "hyperview.in/server/base"
)


type Result struct {
  body        []byte
  contentType string
  err         error
  statusCode  int
  reason      string
}

func NewResult(body []byte, contentType string, err error, statusCode int) Result {
  var body_json map[string]interface{}
  var ret_error error
  var reason string 
  
  if statusCode > 201 && contentType == "application/json" { 
    err = json.Unmarshal(body, &body_json)
    err_string, _ := body_json["error"].(string)
    ret_error = fmt.Errorf(err_string)
    reason = body_json["reason"].(string)

  } 

  if ret_error == nil {
    ret_error = err
  }    
  

  return Result {
    body: body,
    contentType: contentType,
    err: ret_error,
    statusCode: statusCode,
    reason: reason,
  }
}

func (r Result) Body() []byte {
  return r.body
}

func (r Result) JsonBody() (json_body map[string]interface{}, err error) {
  err = json.Unmarshal(r.body, &json_body)
  if err != nil {
    base.Log("Failed to convert body to JSON")
    return nil, err
  }
  return json_body, nil
}

func (r Result) Raw() ([]byte, error) {
  if r.reason != "" {
    return r.body, fmt.Errorf(r.err.Error() + r.reason)
  }
  return r.body, r.err
}

func (r Result) StatusCode(statusCode *int) Result {
  *statusCode = r.statusCode
  return r
}



func (r Result) Error() error {
	return r.err
}