package db

import (
  "fmt"
  "time"
	"database/sql"
  _ "github.com/lib/pq"

)


type DatabaseContext struct {
  Name string
  conn *sql.DB
  LastCall time.Time
}

func NewDatabaseContext(db_name string, db_user string, db_password string) (*DatabaseContext, error) {
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

  return &DatabaseContext{Name: db_name, conn: db, LastCall: time.Now()}, nil
}


func (d *DatabaseContext) Close() {
  defer d.conn.Close()
}

