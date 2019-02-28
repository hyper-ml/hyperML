package rest_client


import (
	"testing"
  "fmt"
  "net/url"
)

func Test_NewClient(t *testing.T) {
  url_object, err := url.Parse("https://google.com")
  if err != nil {
    fmt.Println("failed to generate URL", err)
  }
  
  client, err := NewRESTClient(url_object, "", nil)

  request:= client.Verb("GET")
  result:= request.Do()

  fmt.Println("result", string(result.body))

}
