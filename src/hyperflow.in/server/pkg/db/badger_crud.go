package db

import ( 
  "fmt"
  "encoding/json"
  "github.com/dgraph-io/badger"

)

func (b *BadgerContext) Insert(key string, value interface{}) error{

  if b.KeyExists(key) {
    return fmt.Errorf("[DB.Insert] key already exists in database: %s", key)
  }

  k := []byte(key)

  json_value, err := json.Marshal(value)
  if (err != nil) {
    return err
  }

  err = b.db.Update(func(txn *badger.Txn) error{
    return txn.Set(k, json_value)
  })

  return err
}

func (b *BadgerContext) Get(key string) ([]byte, error) {
  
  var data []byte
  var err error  
  k := []byte(key)

  err = b.db.View(func(tx *badger.Txn) error {
    
    item, err := tx.Get(k)
    if err != nil {
      return ErrRecNotFound{}
    }

    if item == nil {
      return ErrRecNotFound{}
    }

    data, err = item.ValueCopy(nil)
    return err
  })

  if err != nil {
    return nil, err
  }

  if data == nil {
    return data, fmt.Errorf("invalid key: does not exist")
  }

  return data, err 
}


func (b *BadgerContext) Upsert(key string, value interface{}) error{

  if b.KeyExists(key) {
    return b.Update(key, value)
  } else {
    return b.Insert(key, value)
  }

}
func (b *BadgerContext) Update(key string, value interface{}) error {
  k := []byte(key)

  if !b.KeyExists(key) {
    return ErrRecNotFound{}
  }

  json_value, err := json.Marshal(value)
  if (err != nil) {
    return err
  }


  return b.db.Update(func(tx *badger.Txn) error{
    err:= tx.Set(k, json_value)
    return err
  })

}

func (b *BadgerContext) UpdateAndTrack(key string, value interface{}, template interface{}) error {
  
  err := b.Update(key, value)
  if err != nil {
    return err
  }

  if b.listener != nil && template != nil {
    b.listener.TrackEvent(value, template)
  }

  return nil
}


func (b *BadgerContext) KeyExists(key string) (bool) {
  var exsts bool = false
  k := []byte(key)

  err := b.db.View(func(txn *badger.Txn) error {
    v, err := txn.Get(k); 

    if v != nil {
      exsts = true
    } else if err == badger.ErrKeyNotFound {
      return ErrRecNotFound{}
    } else {
      return err
    }
    return nil
  })

  if err != nil {
    return false
  }

  return exsts
}

func (b *BadgerContext) DeleteIfExists(key string) error{
  
  k := []byte(key)

  return b.db.Update(func(tx *badger.Txn) error{
    err := tx.Delete(k)
    return err
  })
}

func (b *BadgerContext) SoftDelete(key string) error {
  var err error
  k := []byte(key)

  txn := b.db.NewTransaction(true)
  defer txn.Discard()

  item, err := txn.Get(k) 
  if err != nil {
    return fmt.Errorf("failed to get value for key %s, err: %v", key, err)
  }

  v, err := item.ValueCopy(nil)
  if err != nil {
    return err
  }

  err = txn.Set([]byte("deleted: "+ key), v)
  if err != nil {
    return err
  }

  err = txn.Delete(k)

  return txn.Commit(func(error){})
}

func (b *BadgerContext) Delete(key string) error{
  return b.DeleteIfExists(key)
}

func (b *BadgerContext) GetListener() ChangeListener{
  return b.listener
}

func (b *BadgerContext) Close() {
  if b.listener!= nil {
    b.listener.Close()
  } 
  defer b.db.Close()
} 
