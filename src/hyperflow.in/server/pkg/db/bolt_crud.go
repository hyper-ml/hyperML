package db

/* -- No Bolt for now -- 


import(
  "fmt"
  "encoding/json"
  bolt "go.etcd.io/bbolt"

)


type BoltContext struct {
  Name string
  db *bolt.DB
  LastCall time.Time
  listener ChangeListener
  // add listener active flag to avoid multiple listener
}

func NewBoltContext(name, datapath string) (DatabaseContext, error) {
  
  db, err := bolt.Open(datapath, 0600, &bolt.Options{Timeout: 5 * time.Second})
  if err != nil {
    return nil, fmt.Errorf("failed to open bolt db %s: %v", datapath, err)
  }

  if err := db.Update(func(tx *bolt.Tx) error {
    _ , err := tx.CreateBucketIfNotExists([]byte(name))
    if err != nil {
      return fmt.Errorf("failed to create bolt db bucket %s: %v", name, err)
    }
    return nil
  }); err != nil {
    return nil, err
  }

  return &BoltContext{
    Name: name,
    db: db,
    LastCall: time.Now(),
    listener: NewChangeListener(25),
  }, nil
}

func (b *BoltContext) Insert(key string, value interface{}) error{

  if b.KeyExists(key) {
    return fmt.Errorf("[DB.Insert] key already exists in database: %s", key)
  }

  k := []byte(key)
  json_value, err := json.Marshal(value)
  if (err != nil) {
    return err
  }

  return b.db.Update(func(tx *bolt.Tx) error{
    bucket := tx.Bucket([]byte(b.Name)) 
    return bucket.Put(k, json_value)
  })

}

func (b *BoltContext) Upsert(key string, value interface{}) error{

  if b.KeyExists(key) {
    return b.Update(key, value)
  } else {
    return b.Insert(key, value)
  }

}

func (b *BoltContext) DeleteIfExists(key string) error{
  
  k := []byte(key)

  return b.db.Update(func(tx *bolt.Tx) error{
    bucket := tx.Bucket([]byte(b.Name)) 
    err := bucket.Delete(k)
    return err
  })
}

// TODO: throw error if rec doesnt exist 

func (b *BoltContext) Delete(key string) error{
  return b.DeleteIfExists(key)
}

func (b *BoltContext) SoftDelete(key string) error {
  
  k := []byte(key)

  tx, err := b.db.Begin(true)
  defer tx.Rollback()
  if err != nil {
    return err
  }

  bucket := tx.Bucket([]byte(b.Name))

  v, _ := b.Get(key) 
  if v != nil {
    err = b.Put([]byte("deleted: "+ key), v)
    if err != nil {
      return err
    }
  }

  return tx.Commit()  
}

func (b *BoltContext) KeyExists(key string) (bool) {
  var exsts bool = false
  k := []byte(key)

  err := b.db.View(func(tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte(b.Name))
    if v := bucket.Get(k); v != nil {
      exsts = true
    }
  })

  if err != nil {
    return false
  }

  return false
}

func (b *BoltContext) Get(key string) ([]byte, error) {
  
  var data []byte
  var err error  
  k = []byte(key)

  err = b.db.View(func (tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte(b.Name))
    data = bucket.Get(k)
  })

  if data == nil {
    return data, fmt.Errorf("invalid key: does not exist")
  }

  return data, err 
}

// 
func (b *BoltContext) Update(key string, value interface{}) error {
  k := []byte(key)

  if !b.KeyExists(key) {
    return ErrRecNotFound{}
  }

  json_value, err := json.Marshal(value)
  if (err != nil) {
    return err
  }


  return b.db.Update(func(tx *bolt.Tx) error{
    bucket := tx.Bucket([]byte(b.Name))
    err:= bucket.Put(k, json_value)
    return err
  })

}

// a bad hack to enable tracking of  updates 
//
func (b *BoltContext) UpdateAndTrack(key string, value interface{}, template interface{}) error {
  
  err := b.Update(key, value)
  if err != nil {
    return err
  }

  if b.listener != nil && template != nil {
    b.listener.TrackEvent(value, template)
  }

  return nil
}
 

func (b *BoltContext) Close() {
  
  if b.listener!= nil {
    b.listener.Close()
  } 
  defer b.db.Close()
}

*/
