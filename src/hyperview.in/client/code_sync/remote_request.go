package code_sync

import (  
  "net/url" 
  "net/http"
  "strings"
  "fmt"
  "path" 
  "io"
)



type HTTPClient interface {
  Do(req *http.Request) (*http.Response, error)
}

type remoteRequest struct {
  client  HTTPClient
  baseURL *url.URL
  verb string 
  subPath string 
  pathPrefix string

  params     url.Values
  headers    http.Header 

  err        error
}

func (r *remoteRequest) setParam(paramName, value string) *remoteRequest {
  if r.params == nil {
    r.params = make(url.Values)
  }
  r.params[paramName] = append(r.params[paramName], value)
  return r
}

func (r *remoteRequest) URL() *url.URL{

  destURL := &url.URL{}
  if r.baseURL != nil {
    *destURL = *r.baseURL
  }

  p := r.pathPrefix
  if len(r.subPath) != 0 {
    p = path.Join(p, strings.ToLower(r.subPath))
  }

  destURL.Path = p

  query := url.Values{}
  for key, values := range r.params {
    for _, value := range values {
      query.Add(key, value)
    }
  }

  /*TODO: if r.timeout != 0 {
    query.Set("timeout", r.timeout.String())
  }*/

  destURL.RawQuery = query.Encode() 

  fmt.Println("print destination URL", destURL)
  return destURL
}

func (r *remoteRequest) SetHeader(key string, values ...string) *remoteRequest {
  if r.headers == nil {
    r.headers = http.Header{}
  }
  r.headers.Del(key)
  for _, value := range values {
    r.headers.Add(key, value)
  }
  return r
}

func (r *remoteRequest) SetVerb(v string) *remoteRequest {
  r.verb = v
  return r
}

//TODO: add context 
func (r *remoteRequest) ReadResponse() (io.ReadCloser, error) {
  client := r.client
  if client == nil {
    client = http.DefaultClient
  }

  url := r.URL().String()

  req, err := http.NewRequest("GET", url, nil)
  req.Header = r.headers

  resp, err := client.Do(req) 
  fmt.Println("response", resp, err)

  if err != nil {
    return nil, err
  }

  return resp.Body, nil
} 
