package rest_client


import (
  "strings" 
  "net/url"
  "net/http"
)


//TODO: add back-off, API Version


type Interface interface {
	Verb(verb string) *Request
  Post() *Request
  Put() *Request
  Get() *Request
  Delete() *Request
}



type RESTClient struct {
  // root URL for all invocations
  base *url.URL
  //path to resource
  apiPath string 
  //TODO: add content config for server communication
  //TODO: add a backoff manager
  //TODO: add throttle
  Client *http.Client
}


//TODO: add max qps, max burst for throttle, rate limiter etc
func NewRESTClient(baseURL *url.URL, apiPath string, client *http.Client) (*RESTClient, error) {
  base := *baseURL
  if !strings.HasSuffix(base.Path, "/") {
    base.Path += "/"
  }
  base.RawQuery = ""
  base.Fragment = "" 

  return &RESTClient{
    base: &base,
    apiPath: apiPath,
    Client: client,
  }, nil
}


func (c *RESTClient) Verb(verb string) *Request {
  if c.Client == nil {
    return NewRequest(nil, verb, c.base, c.apiPath, 0)
  }
  return NewRequest(c.Client, verb, c.base, c.apiPath, c.Client.Timeout)
}

func (c *RESTClient) Post() *Request {
  return c.Verb("POST")
}

func (c *RESTClient) Put() *Request {
  return c.Verb("PUT")
}

func (c *RESTClient) Get() *Request {
  return c.Verb("GET")
}
 

func (c *RESTClient) Delete() *Request {
  return c.Verb("DELETE")
}

func (c *RESTClient) APIVersion() string {
  return "0.1"
}
