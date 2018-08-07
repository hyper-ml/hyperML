package db

import (
  "fmt"
  "encoding/json"
  "database/sql"
  "hyperview.in/server/base"
)

//TODO: Add entity checks 

type Body map[string]interface{}

func (d *DatabaseContext) Insert(key string, value interface{}) error{
	tx, err := d.conn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("INSERT INTO ws_collections(key, value) VALUES ($1, $2)")
  if err != nil {
    return err
  }

  //defer stmt.Close()
  json_value, err := json.Marshal(value)
  if (err != nil) {
    return err
  }

  if _, err = stmt.Exec(key, json_value); err != nil {
    tx.Rollback()
    return err
  }

  // really bad hack to capture flow chagnes 
  return tx.Commit()
}

func (d *DatabaseContext) Upsert(key string, value interface{}) error{

  if d.KeyExists(key) {
    return d.Update(key, value)
  } else {
    return d.Insert(key, value)
  }

}

func (d *DatabaseContext) DeleteIfExists(key string) error{
  tx, err := d.conn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("DELETE FROM ws_collections WHERE key=$1")
  if err != nil {
    return err
  }

  if _, err = stmt.Exec(key); err != nil {
    tx.Rollback()
    return err
  }

  return tx.Commit()
}

// TODO: throw error if rec doesnt exist 

func (d *DatabaseContext) Delete(key string) error{
  tx, err := d.conn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("DELETE FROM ws_collections WHERE key=$1")
  if err != nil {
    return err
  }

  if _, err = stmt.Exec(key); err != nil {
    tx.Rollback()
    return err
  }

  return tx.Commit()
}

func (d *DatabaseContext) SoftDelete(key string) error {
  tx, err := d.conn.Begin()
  if err != nil {
    return err
  }
  trash_key:= "deleted:" + key 

  stmt, err := tx.Prepare("UPDATE ws_collections SET key = $1 WHERE key=$2")
  if err != nil {
    return err
  }

  if _, err = stmt.Exec(trash_key, key); err != nil {
    tx.Rollback()
    return err
  }

  return tx.Commit()
}

func (d *DatabaseContext) KeyExists(key string) (bool) {
  var err error
  var data []byte

  rec, err := d.conn.Query("SELECT 1 FROM ws_collections WHERE key=$1", key)
  
  if err != nil {
    base.Log("Something wrong with database connection. Key Exists Error: %s", err)
    return false
  } 

  defer rec.Close()

  for rec.Next() {
    err = rec.Scan(&data)
  }

  if data != nil {
    base.Log("File key exists. Updating file info")
    return true
  }

  return false
}

func (d *DatabaseContext) Get(key string) ([]byte, error) {
  var data []byte
  var err error 

  rows, err := d.conn.Query("SELECT value FROM ws_collections WHERE key=$1", key)

  switch {
  
  case err == sql.ErrNoRows:
    base.Debug("DatabaseContext.Get(): Found no key in DB", key)
    return nil, ErrRecNotFound{}

  case err != nil:
    fmt.Println("failed to retrive rows", err)
    return nil, err

  default:
  }
  
  defer rows.Close()

  for rows.Next() {
    err = rows.Scan(&data) 
  }
  
  if data == nil {
    return nil, ErrRecNotFound{}
  }

  return data, err
}


// TODO: raises error if key is missing
//
func (d *DatabaseContext) Update(key string, value interface{}) error {
  tx, err := d.conn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("UPDATE ws_collections SET value=$1 WHERE key=$2")
  if err != nil {
    return err
  }

  //defer stmt.Close()
  json_value, err := json.Marshal(value)
  if (err != nil) {
    return err
  }

  if _, err = stmt.Exec(json_value, key); err != nil {
    tx.Rollback()
    return err
  }
 
  return tx.Commit()
}

// a bad hack to enable tracking of  updates 
//
func (d *DatabaseContext) UpdateAndTrack(key string, value interface{}, template interface{}) error {
  tx, err := d.conn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("UPDATE ws_collections SET value=$1 WHERE key=$2")
  if err != nil {
    return err
  }

  //defer stmt.Close()
  json_value, err := json.Marshal(value)
  if (err != nil) {
    return err
  }

  if _, err = stmt.Exec(json_value, key); err != nil {
    tx.Rollback()
    return err
  }
  err = tx.Commit()
  if err != nil {
    return err
  }

  if d.Listener != nil && template != nil {
    d.Listener.TrackEvent(value, template)
  }

  return nil
}

