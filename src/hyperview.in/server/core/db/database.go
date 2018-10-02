package db

import (
  "fmt"
  "time"
	"database/sql"
  _ "github.com/lib/pq"

  "hyperview.in/server/base"

)


type DatabaseContext struct {
  Name string
  conn *sql.DB
  LastCall time.Time
  Listener ChangeListener
}

// TODO: add listener activate flag
func NewDatabaseContext(driver string, name string, user string, pass string) (*DatabaseContext, error) {
  
  switch driver {
    case 'POSTGRES': 
      return NewPostgresContext(name, user, pass)
  }

  return nil, fmt.Errorf("Failed to initiate database context")
}

func NewPostgresContext(name, user, pass string) (*DatabaseContext, error) {
  var err error
  var conn_str string

  if db_password != "" {
    conn_str = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
                db_user, db_password, db_name)
  } else {
    conn_str = fmt.Sprintf("user=%s dbname=%s sslmode=disable",
                db_user, db_name)
  }

  db, err := sql.Open("postgres", conn_str)
  if err != nil {
    return nil, err
  }

  //TODO: add config for listener
  change_lnr := NewChangeListener(25)

  return &DatabaseContext{Name: db_name, conn: db, LastCall: time.Now(), Listener: change_lnr}, nil

}


func (d *DatabaseContext) Close() {
  
  if d.Listener!= nil {
    base.Log("Closing Listener")
    d.Listener.Close()
  } 
  defer d.conn.Close()
}

