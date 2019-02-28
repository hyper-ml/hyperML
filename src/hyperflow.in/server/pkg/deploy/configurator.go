package deploy
import (
  "fmt"
  "bytes"
  "io/ioutil"
  "github.com/ghodss/yaml"

)

const (
  lineSep = "---\n"
)

type Configurator interface{
  AddEntry(item interface{}) error
  ReadCacheBytes()  []byte
  FlushToFile(path string) error
}

type configurator struct {
  b *bytes.Buffer 
}

func NewConfigurator() Configurator {
  b := &bytes.Buffer{}
  return &configurator{b}
}

func (c *configurator) AddEntry(item interface{}) error {
  
  item_bytes, err := yaml.Marshal(item)
  if err != nil {
    return fmt.Errorf("failed to marshal item, err: %v", err)
  }

  if _, err := c.b.Write(item_bytes); err != nil {
    return fmt.Errorf("failed to write marshal config data to buffer")
  } 
  _, err = fmt.Fprintf(c.b, lineSep)
  return err
}

func (c *configurator) ReadCacheBytes() []byte {
  return c.b.Bytes()
}

func (c *configurator) FlushToFile(path string) error {
  if path == "" {
    path = "hfserver.yml"  
  }

  err := ioutil.WriteFile(path, c.ReadCacheBytes(), 0644)
  return err
}