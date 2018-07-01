package db

import (
  "testing"
)


func TestIntegration_dbInsert(t *testing.T) {
  dc, err := NewDatabaseContext(DB_NAME, DB_USER, DB_PASSWORD)
  if err != nil {
    t.Fatalf("Authentication Error in initialization of DB")
  }

  defer dc.Close()

  err= dc.Insert("test-key", Body{"col1": "value1"})

  if err != nil {
    t.Fatalf("Failed to insert a test record in wc_collections")
  }
  err = dc.Delete("test-key")

  if err != nil {
    t.Fatalf("Failed to delete a test-key in wc_collections")
  } 
}