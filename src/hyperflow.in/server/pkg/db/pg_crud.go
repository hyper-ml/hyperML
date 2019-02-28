package db

import ( 
  "encoding/json"
  "database/sql"
  "hyperflow.in/server/pkg/base"
)

//TODO: Add entity checks 


func (p *PGContext) Insert(key string, value interface{}) error{
	tx, err := p.conn.Begin()
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

  return tx.Commit()
}

func (p *PGContext) Upsert(key string, value interface{}) error{

  if p.KeyExists(key) {
    return p.Update(key, value)
  } else {
    return p.Insert(key, value)
  }

}

func (p *PGContext) DeleteIfExists(key string) error{
  tx, err := p.conn.Begin()
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

func (p *PGContext) Delete(key string) error{
  tx, err := p.conn.Begin()
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

func (p *PGContext) SoftDelete(key string) error {
  tx, err := p.conn.Begin()
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

func (p *PGContext) KeyExists(key string) (bool) {
  var err error
  var data []byte

  rec, err := p.conn.Query("SELECT 1 FROM ws_collections WHERE key=$1", key)
  
  if err != nil {
    base.Log("Something wrong with database connection. Key Exists Error: %s", err)
    return false
  } 

  defer rec.Close()

  for rec.Next() {
    err = rec.Scan(&data)
  }

  if data != nil {
    base.Info("Updating file info for key: ", key)
    return true
  }

  return false
}

func (p *PGContext) Get(key string) ([]byte, error) {
  var data []byte
  var err error  
 
  rows, err := p.conn.Query("SELECT value FROM ws_collections WHERE key=$1", key)
  if err != nil {
    if (err == sql.ErrNoRows) {
      return nil, ErrRecNotFound{}
    }  

    return nil, err
  }
  
  defer rows.Close()
  
  for rows.Next() {
    err = rows.Scan(&data) 
  }
  
  if data == nil {
    base.Error("[PGContext.Get] No data found for key", key, data)
    return nil, ErrRecNotFound{}
  }

  return data, err
}


// TODO: raises error if key is missing
//
func (p *PGContext) Update(key string, value interface{}) error {
  tx, err := p.conn.Begin()
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
func (p *PGContext) UpdateAndTrack(key string, value interface{}, template interface{}) error {
  tx, err := p.conn.Begin()
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

  if p.Listener != nil && template != nil {
    p.Listener.TrackEvent(value, template)
  }

  return nil
}

func (p *PGContext) GetListener() ChangeListener{
  return p.Listener
}

func (p *PGContext) Close() {
  
  if p.Listener!= nil {
    p.Listener.Close()
  } 
  defer p.conn.Close()
} 


