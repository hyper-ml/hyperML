package db

import (
  "fmt"
  "time"
	"database/sql"
  _ "github.com/lib/pq"
  "github.com/dgraph-io/badger"

  "hyperflow.in/server/pkg/config"

)

type Body map[string]interface{}

type DatabaseContext interface {
  Get(key string) ([]byte, error)
  Insert(key string, value interface{}) error
  Update(key string, value interface{}) error
  UpdateAndTrack(key string, value interface{}, template interface{}) error
  Upsert(key string, value interface{}) error

  DeleteIfExists(key string) error
  Delete(key string) error
  SoftDelete(key string) error
  KeyExists(key string) (bool)
  GetListener() ChangeListener
  Close()
}
 
func NewDatabaseContext(c *config.DBConfig) (DatabaseContext, error) {
  
  switch c.Driver {
    case config.Postgres: 
      return NewPGContext(c.Name, c.User, c.Pass)
    case config.Badger:
      return NewBadgerContext(c.Name, c.DataDirPath)
  }

  return nil, fmt.Errorf("Failed to initiate database context")
}

type BadgerContext struct {
  Name string
  db *badger.DB
  LastCall time.Time
  listener ChangeListener
  // add listener active flag to avoid multiple listener
}

func NewBadgerContext(name, datapath string) (DatabaseContext, error) {
  opts := badger.DefaultOptions
  opts.Dir = datapath
  opts.ValueDir = datapath
  db, err := badger.Open(opts)
  if err != nil {
    return nil, err
  }

  return &BadgerContext{
    Name: name,
    db: db,
    LastCall: time.Now(),
    listener: NewChangeListener(25),
  }, nil

}


type PGContext struct {
  Name string
  conn *sql.DB
  LastCall time.Time
  Listener ChangeListener
  // add listener active flag to avoid multiple listener
}

func NewPGContext(name, user, pass string) (DatabaseContext, error) {
  var err error
  var conn_str string

  if pass != "" {
    conn_str = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
                user, pass, name)
  } else {
    conn_str = fmt.Sprintf("user=%s dbname=%s sslmode=disable",
                user, name)
  }

  db, err := sql.Open("postgres", conn_str)
  if err != nil {
    return nil, err
  }

  //TODO: add config for listener
  change_lnr := NewChangeListener(25)

  return &PGContext{Name: name, conn: db, LastCall: time.Now(), Listener: change_lnr}, nil

}

