package rest_client

import ( 
  "fmt"
  "strconv"
  "net/url"
  "encoding/json"

  "github.com/hyper-ml/hyperml/server/pkg/base"
)

type Result struct {
  url         *url.URL
  body        []byte
  contentType string
  err         error
  statusCode  int
  reason      string
}

func NewResult(url *url.URL, body []byte, contentType string, err error, statusCode int) Result {
  var body_json map[string]interface{}
  var ret_error error
  var reason string 
  //base.Debug("[result.NewResult] statusCode: ", statusCode)

  if statusCode > 201 && contentType == "application/json" { 
    err = json.Unmarshal(body, &body_json)

    err_string, _ := body_json["error"].(string)
    //base.Debug("[result.NewResult] err_string: ", err_string)
    if err_string == "" {
      err_string = "http Error: " + strconv.Itoa(statusCode)
    }

    ret_error = fmt.Errorf(err_string)

    reason_bytes, _ := body_json["reason"]
    if reason_bytes != nil {
      reason = reason_bytes.(string)
    }
  } 

  if ret_error == nil {
    ret_error = err
  }    
  //base.Debug("[result.NewResult] ret_error: ", ret_error)

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