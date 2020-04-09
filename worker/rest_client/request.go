package rest_client

import (
  "fmt"
  "bytes"  
  "strings"  
  "io/ioutil"
  "net/http"
  "net/url"
  "time"
  "io"
  "path"
  "context"
  "strconv"
  "github.com/hyper-ml/hyperml/server/pkg/base"
)


type HTTPClient interface {
  Do(req *http.Request) (*http.Response, error)
}

//TODO: add throttle logging


type Request struct {
	client HTTPClient
  verb   string
  baseURL     *url.URL

  pathPrefix string
  subpath    string
  params     url.Values
  headers    http.Header
  
  timeout time.Duration

  err  error
  body io.Reader

  // use this for per request cancels
  ctx context.Context

  //TODO: add throttle
}


//ToDO: result into
// TODO: add config param
func NewRequest(client HTTPClient, verb string, baseURL *url.URL, apiPath string, subPath string, timeout time.Duration) *Request {

  pathPrefix := "/"
  if baseURL != nil {
    pathPrefix = path.Join(pathPrefix, baseURL.Path)
  }

  if apiPath != "" {
    pathPrefix = path.Join(pathPrefix, apiPath)
  }

  if subPath != "" {
    pathPrefix = path.Join(pathPrefix, subPath)
  }

  r := &Request{
    client:      client,
    verb:        verb,
    baseURL:     baseURL,
    pathPrefix:  pathPrefix,
  }
  r.SetHeader("Accept", "application/json")

  // TODO: add default content and accept content type here
  return r
} 

func (r *Request) request(fn func(*http.Request, *http.Response)) error {
  client := r.client
  if client == nil {
    client = http.DefaultClient
  }

  maxRetries := 1
  retries := 0
  for {
    url := r.URL().String()
    req, err := http.NewRequest(r.verb, url, r.body)
    if err != nil {
      return err
    }

    if r.timeout > 0 {
      if r.ctx == nil {
        r.ctx = context.Background()
      }
      var cancelFn context.CancelFunc
      r.ctx, cancelFn = context.WithTimeout(r.ctx, r.timeout)
      defer cancelFn()
    }

    if r.ctx != nil {
      req = req.WithContext(r.ctx)
    }
    req.Header = r.headers

    resp, err := client.Do(req) 

    if err != nil {

      if !IsConnectionReset(err) || r.verb != "GET" {
        return err
      }
      resp = &http.Response{
        StatusCode: http.StatusInternalServerError,
        Header:     http.Header{"Retry-After": []string{"1"}},
        Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
      }
    }  

    done := func() bool {
      
      retries++
      if _, wait := checkWait(resp); wait && retries < maxRetries {
        if seeker, ok := r.body.(io.Seeker); ok && r.body != nil {
          _, err := seeker.Seek(0, 0)

          if err != nil {
            base.Log("Could not retry request, can't Seek() back to beginning of body for %T", r.body)
            fn(req, resp)
            return true
          }
        }
        return false
      }
      fn(req, resp)
      return true
    }()

    if done {
      return nil
    }
  }
}


func checkWait(resp *http.Response) (int, bool) {
  switch r := resp.StatusCode; {
  // any 500 error code and 429 can trigger a wait
  case r == http.StatusTooManyRequests, r >= 500:
  default:
    return 0, false
  }
  i, ok := retryAfterSeconds(resp)
  return i, ok
}

func retryAfterSeconds(resp *http.Response) (int, bool) {
  if h := resp.Header.Get("Retry-After"); len(h) > 0 {
    if i, err := strconv.Atoi(h); err == nil {
      return i, true
    }
  }
  return 0, false
}


func (r *Request) Do() Result {

  var result Result
  err := r.request(func(req *http.Request, resp *http.Response) {
    result = r.processJsonResponse(resp, req) 
  })
  if err != nil {
    return Result{err: err}
  }
  return result
}


func (r *Request) processJsonResponse(resp *http.Response, req *http.Request) Result {
  var body []byte
  var err error
  url :=  r.URL()
  if resp.Body != nil {
    body, err = ioutil.ReadAll(resp.Body)
    //base.Debug("[Request.processJsonResponse] Result: ", string(body), err)

    if err != nil {
      base.Log("[Request.processJsonResponse] HTTP request Failed: ", err)
      return Result{url: url, err: err}
    }

     //TODO: change content type based on response
    return NewResult(url, body, "application/json" , err, resp.StatusCode)
  }
  return Result{}
}



// Param creates a query parameter with the given string value.
func (r *Request) Param(paramName, s string) *Request {
  if r.err != nil {
    return r
  }
  return r.setParam(paramName, url.PathEscape(s))
}

func (r *Request) setParam(paramName, value string) *Request {
  if r.params == nil {
    r.params = make(url.Values)
  }
  r.params[paramName] = append(r.params[paramName], value)
  return r
}

func (r *Request) PrintParams() error {
  base.Log("Request params:", r.params)
  return nil
}


func (r *Request) SetHeader(key string, values ...string) *Request {
  if r.headers == nil {
    r.headers = http.Header{}
  }
  r.headers.Del(key)
  for _, value := range values {
    r.headers.Add(key, value)
  }
  return r
}

func (r *Request) SetBodyReader(t io.Reader) *Request {
  r.body = t
  return r
}

/*func (r *Request) Body(obj interface{}) *Request {
  if r.err != nil {
    return r
  }

  switch t := obj.(type) {
  case string:
    r.body = strings.NewReader(obj)
  case []byte:
    r.body = bytes.NewReader(t)
  case io.Reader:
    r.body = t
  default:
    r.err = fmt.Errorf("unknown type used for body: %+v", obj)
  }
  return r
}*/

func (r *Request) URL() *url.URL {
  p := r.pathPrefix
  if len(r.subpath) != 0 {
    p = path.Join(p, strings.ToLower(r.subpath))
  }
  finalURL := &url.URL{}
  if r.baseURL != nil {
    *finalURL = *r.baseURL
  }
  
  finalURL.Path = p
  query := url.Values{}
  for key, values := range r.params {
    for _, value := range values {
      query.Add(key, value)
    }
  }
  if r.timeout != 0 {
    query.Set("timeout", r.timeout.String())
  }

  finalURL.RawQuery = query.Encode()
  return finalURL
}


func (r *Request) Context(ctx context.Context) *Request {
  r.ctx = ctx
  return r
}

/*func (r *Request) Into(obj interface{}) (interface{}, error) {
  err := json.Unmarshal(r.body, obj)
  if err != nil {
    return nil, err
  }
  return obj, nil
}*/


// Raw non-JSON outcome
func (r *Request) DoRaw() ([]byte, error) {

  var result Result
  err := r.request(func(req *http.Request, resp *http.Response) {
    result.body, result.err = ioutil.ReadAll(resp.Body)
    
    if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent {
      result.err = fmt.Errorf("HTTP Failure: %s", result.body) 
      // TODO: handle un strcutured errors
    }

  })
  if err != nil {
    return nil, err
  }
  return result.body, result.err
}

// read raw output from server
func (r *Request) ReadResponse() (io.ReadCloser, error) {
  client := r.client
  if client == nil {
    client = http.DefaultClient
  }

  url := r.URL().String()

  req, err := http.NewRequest(r.verb, url, nil)
  req.Header = r.headers

  resp, err := client.Do(req)  

  if err != nil {
    return nil, err
  }

  return resp.Body, nil
} 

