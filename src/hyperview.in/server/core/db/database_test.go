package db

import (
  "fmt"
  "testing"
)

const (
  DB_USER = "apple"
  DB_PASSWORD = ""
  DB_NAME = "amp_db"
)

func TestIntegration_dbConnect(t *testing.T) {
	dc, err := NewDatabaseContext(DB_NAME, DB_USER, DB_PASSWORD)
  if err != nil {
    t.Fatalf("Authentication Error in initialization of DB")
  }

  defer dc.Close()

  _, err = dc.conn.Query("SELECT 'Hello world'")

  if err != nil {
    fmt.Println("error", err)
    t.Fatalf("Unexpected Error when running select query test")
  }
}

